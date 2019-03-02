package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

var (
	//debug      = log.New(os.Stdout /* ioutil.Discard */, "[DEBUG]", log.Lshortfile)
	allReceivers = MakeAllReceivers()
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

// A list of receivers that are listening on a pipe

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

// the number of receivers in the list
func (mw *ReceiverList) Count() int {
	return len(mw.writers)
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

// A map of receiver lists by a key

type AllReceivers struct {
	receiverLists map[string]ReceiverList
}

// add a new receiver to a key
func (ar AllReceivers) Add(key string, receiver Receiver) {
	receiverlist, exists := ar.receiverLists[key]
	if !exists {
		// create the receiver list if it doesn't exist
		receiverlist = MakeReceiverList()
		ar.receiverLists[key] = receiverlist
	}
	receiverlist.Add(receiver)
}

// remove a receiver from a key
func (ar AllReceivers) Delete(key string, receiver Receiver) {
	receiverlist, exists := ar.receiverLists[key]
	if !exists {
		return
	}

	receiverlist.Delete(receiver)
	if receiverlist.Count() < 1 {
		// remove the receiver list from the map if it is empty
		delete(ar.receiverLists, key)
	}
}

// find a receiverlist by a key
func (ar AllReceivers) Find(key string) (ReceiverList, bool) {
	receiverList, exists := allReceivers.receiverLists[key]
	return receiverList, exists
}

func (ar AllReceivers) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d keys\n", len(ar.receiverLists)))
	for key, receiverList := range ar.receiverLists {
		sb.WriteString(fmt.Sprintf("%s : %d\n", key, receiverList.Count()))
	}
	return sb.String()
}

func MakeAllReceivers() AllReceivers {
	return AllReceivers{
		receiverLists: make(map[string]ReceiverList),
	}
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

var keyPath = regexp.MustCompile("^/([a-z0-9]+)$")

// send or receive based on http method
func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		home(w, r)
		return
	}
	if r.URL.Path == "/favicon.ico" {
		return
	}

	// /<key>
	m := keyPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return
	}
	key := m[1]

	if r.Method == "GET" {
		recv(w, r, key)
		return
	}
	if r.Method == "POST" || r.Method == "PUT" {
		send(w, r, key)
		return
	}
	http.Error(w, "Invalid Method", http.StatusNotFound)
}

// handler to show basic information on the home page
func home(w http.ResponseWriter, r *http.Request) {
	newkey := randASCIIBytes(8)
	url := fmt.Sprintf("http://%s/%s", r.Host, newkey)
	receive := fmt.Sprintf("curl -s %s", url)
	send := fmt.Sprintf("curl -T- -s %s", url)

	fmt.Fprintln(w, "# pipe url")
	fmt.Fprintln(w, url)
	fmt.Fprintln(w, "# receive from pipe")
	fmt.Fprintln(w, receive)
	fmt.Fprintln(w, "# send to pipe")
	fmt.Fprintln(w, send)
}

// receive data from any senders
func recv(w http.ResponseWriter, r *http.Request, key string) {
	// this is required so that data is streamed back to the client
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// this is used to flush output back to the client as it is received
	flusher, _ := w.(http.Flusher)
	flusher.Flush()

	// store the active streams by key so that data can be sent by another request
	receiver := MakeReceiver(w, flusher)
	allReceivers.Add(key, receiver)
	defer allReceivers.Delete(key, receiver)

	done := false
	for done == false {
		select {
		// the receiver disconnected before completion
		case <-r.Context().Done():
			done = true
		// a sender completed a transfer and closed the stream (EOF received)
		case <-receiver.CloseNotify():
			done = true
		}
	}
}

// send data to any connected receivers
func send(w http.ResponseWriter, r *http.Request, key string) {
	// upload size limit
	r.Body = http.MaxBytesReader(w, r.Body, 5*1024*1024)

	// Look to see if there are any receivers attached to this key
	receiverList, exists := allReceivers.Find(key)
	if !exists {
		return
	}

	// copy the body to any listening receivers (see Receivers.Write)
	_, err := io.Copy(receiverList, r.Body)

	// if the copy made it all the way to EOF, close the receivers
	if err == nil {
		receiverList.Close()
	}
}

func main() {
	http.HandleFunc("/", handler)

	// Accept a command line flag "-httpaddr :8080"
	// This flag tells the server the http address to listen on
	httpaddr := flag.String("httpaddr", "localhost:8080",
		"the address/port to listen on for http \n"+
			"use :<port> to listen on all addresses\n")

	flag.Parse()

	log.Println("Listening on http:", *httpaddr)
	log.Fatal(http.ListenAndServe(*httpaddr, nil))
}
