package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
)

const (
	maxUploadMb = 64
	keySize     = 8
)

var (
	baseUrl  = "http://localhost:8080/"
	keyRegex = regexp.MustCompile("^/([a-zA-Z0-9]+)$")
	allPipes = MakePipeCollection()
)

// Handlers

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
	m := keyRegex.FindStringSubmatch(r.URL.Path)
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
	newkey := randKey(keySize)
	url := fmt.Sprintf("%s%s", baseUrl, newkey)
	receive := fmt.Sprintf("curl -s %s", url)
	send := fmt.Sprintf("curl -T- -s %s", url)

	fmt.Fprintf(w, `PIPE TO ME
==========

Your randomly generated pipe address:
	%s

Input example:
	browse to (chrome, firefox): %s
	%s
	hello world<enter>

Pipe example:
	separate terminal: %s
	echo hello world | %s

File transfer example:
	%s > output.txt
	cat input.txt | %s

Watch log example:
	browse to (chrome, firefox): %s
	tail -f logfile | %s

Data is not buffered or stored in any way.
- If data is sent to the pipe when no receivers are listening, 
  it will be dropped and is not retrievable.
- Data is also not retrievable after it has been delivered.

Maximum upload size: %d MB
Not allowed: anything illegal, malicious, inappropriate, etc.

This is a personal project and makes no guarantees on:
	reliability, performance, privacy, etc.

Demo: https://raw.githubusercontent.com/jpschroeder/pipe-to-me/master/demo.gif
Source: https://github.com/jpschroeder/pipe-to-me
`, url,
		url, send, /* input example */
		receive, send, /* pipe example */
		receive, send, /* file transfer example */
		url, send, /* watch log example */
		maxUploadMb)
}

func stats(w http.ResponseWriter, r *http.Request) {
	global := allPipes.GlobalStats()
	gstr := fmt.Sprintf(`
Total Pipes: 		%d
Total Receivers: 	%d
Total Senders: 		%d
Total Sent: 		%d bytes
`, global.PipeCount, global.ReceiverCount, global.SenderCount, global.BytesSent)

	active := allPipes.ActiveStats()
	astr := fmt.Sprintf(`
Connected Pipes: 	%d
Connected Receivers: 	%d
Connected Senders: 	%d
Connected Sent: 	%d bytes
`, active.PipeCount, active.ReceiverCount, active.SenderCount, active.BytesSent)

	fmt.Fprintf(w, "STATS\n=====\n%s%s\n", astr, gstr)
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
	defer allPipes.RemoveReceiver(key, receiver)

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
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadMb*1024*1024)

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
	// Accept a command line flag "-httpaddr :8080"
	// This flag tells the server the http address to listen on
	httpaddr := flag.String("httpaddr", "localhost:8080",
		"the address/port to listen on for http \n"+
			"use :<port> to listen on all addresses\n")

	// Accept a command line flag "-baseurl https://mysite.com/"
	baseurl := flag.String("baseurl", baseUrl,
		"the base url of the service \n")

	flag.Parse()

	baseUrl = *baseurl

	http.HandleFunc("/stats", stats)
	http.HandleFunc("/", handler)

	log.Println("Listening on http:", *httpaddr)
	log.Fatal(http.ListenAndServe(*httpaddr, nil))
}
