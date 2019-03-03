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

var allPipes = MakePipeCollection()

// Hold the information for a single receiver

// a writer that is automatically flushed back to the receiver client
// and a notification channel when it is closed
type Receiver struct {
	writer  io.Writer
	flusher http.Flusher
	done    chan bool
}

// write a single received buffer to the receiver and flush it back to the client
func (r Receiver) Write(p []byte) (n int, err error) {
	n, err = r.writer.Write(p)
	r.flusher.Flush()
	return
}

// close the receiver. flush it one last time and notify that it is closed
func (r Receiver) Close() error {
	r.flusher.Flush()
	r.done <- true
	return nil
}

// a notification channel that will tell when the reciever has been closed
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

// Hold the information for a single pipe

type Pipe struct {
	// a list of receivers that are listening on a pipe
	// allow receivers to be added an removed dynamically
	receivers map[io.WriteCloser]bool
	senders   int
	bytes     int
}

// add a new receiver listening on the pipe
func (p *Pipe) AddReceiver(w io.WriteCloser) {
	p.receivers[w] = true
}

// remove a previously added receiver
func (p *Pipe) RemoveReceiver(w io.WriteCloser) {
	delete(p.receivers, w)
}

// the number of receivers on the pipe
func (p Pipe) ReceiverCount() int {
	return len(p.receivers)
}

// add a new sender connected to send data on the pipe (informational)
func (p *Pipe) AddSender() {
	p.senders++
}

// remove a sender connected to the pipe (informational)
func (p *Pipe) RemoveSender() {
	p.senders--
}

// the number of senders on the pipe
func (p Pipe) SenderCount() int {
	return p.senders
}

// the number of bytes sent through the pipe
func (p Pipe) BytesSent() int {
	return p.bytes
}

// write the buffer to all registered receivers
func (p *Pipe) Write(buffer []byte) (int, error) {
	for receiver := range p.receivers {
		// errors from one of the receivers shouldn't affect any others
		receiver.Write(buffer)
	}
	p.bytes += len(buffer)
	return len(buffer), nil
}

// close all of the registered receivers
func (p Pipe) Close() error {
	for receiver := range p.receivers {
		// errors from one of the receivers shouldn't affect any others
		receiver.Close()
	}
	return nil
}

func (p Pipe) String() string {
	return fmt.Sprintf("%d receivers | %d senders | %d bytes\n",
		p.ReceiverCount(),
		p.SenderCount(),
		p.BytesSent())
}

func MakePipe() *Pipe {
	return &Pipe{
		receivers: make(map[io.WriteCloser]bool),
		senders:   0,
	}
}

// A map of pipes partitioned by a key

type PipeCollection struct {
	// pipe key -> Pipe
	pipes map[string]*Pipe
}

// find a pipe or create one if it doesn't exist
func (pc *PipeCollection) FindOrCreatePipe(key string) *Pipe {
	pipe, exists := pc.pipes[key]
	if !exists {
		pipe = MakePipe()
		pc.pipes[key] = pipe
	}
	return pipe
}

// delete the pipe if it has no attached receivers
func (pc *PipeCollection) DeletePipeIfEmpty(key string, pipe *Pipe) {
	if pipe.ReceiverCount() < 1 && pipe.SenderCount() < 1 {
		delete(pc.pipes, key)
	}
}

// add a new receiver to a pipe - create the pipe if it doesn't exist
func (pc *PipeCollection) AddReceiver(key string, receiver Receiver) {
	pipe := pc.FindOrCreatePipe(key)
	pipe.AddReceiver(receiver)
}

// remove a receiver from a pipe - remove the pipe if its empty
func (pc *PipeCollection) DeleteReceiver(key string, receiver Receiver) {
	pipe, exists := pc.pipes[key]
	if !exists {
		return
	}
	pipe.RemoveReceiver(receiver)
	pc.DeletePipeIfEmpty(key, pipe)
}

// add a new sender to a pipe - create the pipe if it doesn't exist
func (pc *PipeCollection) AddSender(key string) *Pipe {
	pipe := pc.FindOrCreatePipe(key)
	pipe.AddSender()
	return pipe
}

// remove a sender from the pipe - remove the pipe if its empty
func (pc *PipeCollection) RemoveSender(key string, pipe *Pipe) {
	pipe.RemoveSender()
	pc.DeletePipeIfEmpty(key, pipe)
}

func (pc PipeCollection) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d keys\n", len(pc.pipes)))
	for key, pipe := range pc.pipes {
		sb.WriteString(fmt.Sprintf("%s: %s", key, pipe.String()))
	}
	return sb.String()
}

func MakePipeCollection() PipeCollection {
	return PipeCollection{
		pipes: make(map[string]*Pipe),
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

// the root http handler
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

// handler that generates a new key and gives the user information on it
func home(w http.ResponseWriter, r *http.Request) {
	newkey := randASCIIBytes(12)
	url := fmt.Sprintf("http://%s/%s", r.Host, newkey)
	receive := fmt.Sprintf("curl -s %s", url)
	send := fmt.Sprintf("curl -T- -s %s", url)

	fmt.Fprintln(w, "# pipe url")
	fmt.Fprintln(w, url+"\n")
	fmt.Fprintln(w, "# receive from pipe")
	fmt.Fprintln(w, receive+"\n")
	fmt.Fprintln(w, "# send to pipe")
	fmt.Fprintln(w, send+"\n")
}

func stats(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, allPipes.String())
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
	allPipes.AddReceiver(key, receiver)
	defer allPipes.DeleteReceiver(key, receiver)

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
	pipe := allPipes.AddSender(key)
	defer allPipes.RemoveSender(key, pipe)

	// copy the body to any listening receivers (see Receivers.Write)
	_, err := io.Copy(pipe, r.Body)

	// if the copy made it all the way to EOF, close the receivers
	if err == nil {
		pipe.Close()
	}
}

func main() {
	http.HandleFunc("/stats", stats)
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
