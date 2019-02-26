package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
)

var (
	//debug      = log.New(ioutil.Discard /* os.Stdout */, "[DEBUG]", log.Lshortfile)
	allReceivers = MakeReceiverList()
)

// Hold the information for a single receiver

// a writer that is automatically flushed back to the client
// and a notification channel when it is closed
type Receiver struct {
	writer  io.Writer
	flusher http.Flusher
	done    chan bool
}

// write a single received buffer to the writer and flush it back to the client
func (r Receiver) Write(p []byte) (n int, err error) {
	n, err = r.writer.Write(p)
	r.flusher.Flush()
	return
}

// close the writer. flush it one last time and notify that it is closed
func (r Receiver) Close() error {
	r.flusher.Flush()
	r.done <- true
	return nil
}

// a notification channel that will tell when the writer has been closed
func (r Receiver) CloseNotify() <-chan bool {
	return r.done
}

func MakeReceiver(w io.Writer, f http.Flusher) Receiver {
	return Receiver{
		writer:  w,
		flusher: f,
		done:    make(chan bool),
	}
}

// Combine multiple writers together

// allow writers to be added an removed dynamically
type ReceiverList struct {
	writers map[io.WriteCloser]bool
}

// add a new writer to be included in the combined writer
func (mw *ReceiverList) Add(w io.WriteCloser) {
	mw.writers[w] = true
}

// remove a previously entered writer
func (mw *ReceiverList) Delete(w io.WriteCloser) {
	delete(mw.writers, w)
}

// write the buffer to all registered writers
func (mw ReceiverList) Write(p []byte) (int, error) {
	for writer := range mw.writers {
		// errors from one of the writers shouldn't affect any others
		writer.Write(p)
	}
	return len(p), nil
}

// close all of the registered writers
func (mw ReceiverList) Close() error {
	for writer := range mw.writers {
		// errors from one of the writers shouldn't affect any others
		writer.Close()
	}
	return nil
}

func MakeReceiverList() ReceiverList {
	allWriters := make(map[io.WriteCloser]bool)
	return ReceiverList{allWriters}
}

// Utilities

const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"

func randASCIIBytes(n int) []byte {
	output := make([]byte, n)
	// We will take n bytes, one byte for each character of output.
	randomness := make([]byte, n)
	// read all random
	_, err := rand.Read(randomness)
	if err != nil {
		panic(err)
	}
	l := len(letterBytes)
	// fill output
	for pos := range output {
		// get random item
		random := uint8(randomness[pos])
		// random % 64
		randomPos := random % uint8(l)
		// put into output
		output[pos] = letterBytes[randomPos]
	}
	return output
}

// Handlers

// handler to show basic information on the home page
func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
# create new pipe and receive
curl -s http://localhost:8080/pipe | curl -K-

# send to pipe
echo something | curl -T- http://%s/pipe/<key>

# receive from pipe
curl -s http://%s/pipe/<key>

# both
cloud() { test -t 0 && curl -s http://%s/pipe || curl -T- http://%s/pipe; }
	`, r.Host, r.Host, r.Host, r.Host)
}

// send or receive based on http method
func pipe(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("http://%s%s", r.Host, r.RequestURI)
	w.Header().Set("Send-To-Pipe", "curl -T- -s "+url)
	w.Header().Set("Receive-From-Pipe", "curl -s "+url)
	w.Header().Set("New-Pipe-Receive", fmt.Sprintf("curl -s http://%s/pipe | curl -K-", r.Host))
	//w.Header().Set("Content-Location", url)

	if r.Method == "GET" {
		recv(w, r)
	} else if r.Method == "POST" || r.Method == "PUT" {
		send(w, r)
	} else {
		http.Error(w, "Invalid Method", http.StatusNotFound)
	}
}

// receive data from any senders
func recv(w http.ResponseWriter, r *http.Request) {
	// this is required so that data is streamed back to the client
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// this is used to flush output back to the client as it is received
	flusher, _ := w.(http.Flusher)
	// this detects client disconnects
	closer, _ := w.(http.CloseNotifier)

	// store the active streams so that data can be sent by another request
	receiver := MakeReceiver(w, flusher)

	flusher.Flush()

	allReceivers.Add(receiver)
	defer allReceivers.Delete(receiver)

	done := false
	for done == false {
		select {
		// the receiver disconnected before completion
		case <-closer.CloseNotify():
			done = true
		// a sender completed a transfer and closed the stream (EOF received)
		case <-receiver.CloseNotify():
			done = true
		}
	}
}

// send data to any connected receivers
func send(w http.ResponseWriter, r *http.Request) {
	// upload size limit
	r.Body = http.MaxBytesReader(w, r.Body, 5*1024*1024)

	// copy the body to any listening receivers (see Receivers.Write)
	_, err := io.Copy(allReceivers, r.Body)

	// if the copy made it all the way to EOF, close the receivers
	if err == nil {
		allReceivers.Close()
	}
}

func new(w http.ResponseWriter, r *http.Request) {
	newkey := randASCIIBytes(8)
	url := fmt.Sprintf("http://%s/pipe/%s", r.Host, newkey)
	curlconfig := fmt.Sprintf("url=\"%s\"\nsilent\ndump-header=\"/dev/tty\"\n", url)
	w.Header().Set("Send-To-Pipe", "curl -T- -s "+url)
	w.Header().Set("Receive-From-Pipe", "curl -s "+url)
	w.Header().Set("New-Pipe-Receive", fmt.Sprintf("curl -s http://%s/pipe | curl -K-", r.Host))
	//w.Header().Set("Location", url)
	//w.WriteHeader(http.StatusFound)
	fmt.Fprintf(w, curlconfig)
}

func main() {
	http.HandleFunc("/pipe", new)
	http.HandleFunc("/pipe/", pipe)
	http.HandleFunc("/", home)

	// Accept a command line flag "-httpaddr :8080"
	// This flag tells the server the http address to listen on
	httpaddr := flag.String("httpaddr", "localhost:8080",
		"the address/port to listen on for http \n"+
			"use :<port> to listen on all addresses\n")

	flag.Parse()

	log.Println("Listening on http:", *httpaddr)
	log.Fatal(http.ListenAndServe(*httpaddr, nil))
}
