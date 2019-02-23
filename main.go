package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
)

var (
	//debug      = log.New(ioutil.Discard /* os.Stdout */, "[DEBUG]", log.Lshortfile)
	allReceivers = MakeMultiWriteCloser()
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
type MultiWriteCloser struct {
	writers map[io.WriteCloser]bool
}

// add a new writer to be included in the combined writer
func (mw *MultiWriteCloser) Add(w io.WriteCloser) {
	mw.writers[w] = true
}

// remove a previously entered writer
func (mw *MultiWriteCloser) Delete(w io.WriteCloser) {
	delete(mw.writers, w)
}

// write the buffer to all registered writers
func (mw MultiWriteCloser) Write(p []byte) (int, error) {
	for writer := range mw.writers {
		// errors from one of the writers shouldn't affect any others
		writer.Write(p)
	}
	return len(p), nil
}

// close all of the registered writers
func (mw MultiWriteCloser) Close() error {
	for writer := range mw.writers {
		// errors from one of the writers shouldn't affect any others
		writer.Close()
	}
	return nil
}

func MakeMultiWriteCloser() MultiWriteCloser {
	allWriters := make(map[io.WriteCloser]bool)
	return MultiWriteCloser{allWriters}
}

// Handlers

// handler to show basic information on the home page
func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
# receiver
curl -s http://%s/pipe

# sender
echo something | curl -T- http://%s/pipe

# both
cloud() { test -t 0 && curl -s http://%s/pipe || curl -T- http://%s/pipe; }
	`, r.Host, r.Host, r.Host, r.Host)
}

// send or receive based on http method
func pipe(w http.ResponseWriter, r *http.Request) {
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

	// this is used to flush output back to the client as it is received
	flusher, _ := w.(http.Flusher)
	// this detects client disconnects
	closer, _ := w.(http.CloseNotifier)

	// store the active streams so that data can be sent by another request
	receiver := MakeReceiver(w, flusher)

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

func main() {
	http.HandleFunc("/pipe", pipe)
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
