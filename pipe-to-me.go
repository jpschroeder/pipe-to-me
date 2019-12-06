package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"text/template"
	"time"
)

const (
	maxUploadMb = 64
	keySize     = 8
)

// Handlers

type server struct {
	allPipes  PipeCollection
	baseURL   string
	maxID     int
	templates *template.Template
}

var keyRegex = regexp.MustCompile("^/([a-zA-Z0-9]+)$")

type params struct {
	key         string
	id          int    // unique id for this request
	failure     bool   // failure mode will not allow a connection if there is no one on the other end
	block       bool   // block mode will not receive data until there is a connection on the other end
	interactive bool   // interactive mode will send notifications down the pipe on connect/disconnect
	username    string // username passed via basic auth or "" if empty
}

// the root http handler
func (s *server) handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		s.home(w, r)
		return
	}
	if r.URL.Path == "/favicon.ico" {
		return
	}
	if r.URL.Path == "/robots.txt" {
		fmt.Fprintf(w, "User-agent: *\nDisallow: /")
		return
	}
	if r.URL.Path == "/new" {
		fmt.Fprintf(w, "%s%s", s.baseURL, randKey(keySize))
		return
	}
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT")
		return
	}

	params := parseParams(r)
	if params == nil {
		http.NotFound(w, r)
		return
	}
	s.maxID++
	params.id = s.maxID

	if r.Method == "GET" {
		s.recv(w, r, params)
		return
	}
	if r.Method == "POST" || r.Method == "PUT" {
		s.send(w, r, params)
		return
	}
	http.Error(w, "Invalid Method", http.StatusNotFound)
}

func parseParams(r *http.Request) *params {
	// /<key>
	m := keyRegex.FindStringSubmatch(r.URL.Path)
	if m == nil {
		return nil
	}
	key := m[1]
	query := r.URL.Query()
	exists := func(p string) bool {
		return len(query.Get(p)) > 0
	}
	username, _, _ := r.BasicAuth()
	if len(username) == 0 {
		username = query.Get("user")
	}
	return &params{
		key:         key,
		failure:     exists("f") || exists("fail") || query.Get("mode") == "fail",
		block:       exists("b") || exists("block") || query.Get("mode") == "block",
		interactive: exists("i") || exists("interactive"),
		username:    username,
	}
}

// handler that generates a new key and gives the user information on it
func (s *server) home(w http.ResponseWriter, r *http.Request) {
	newkey := randKey(keySize)
	data := struct {
		URL         string
		MaxUploadMb int
	}{
		URL:         fmt.Sprintf("%s%s", s.baseURL, newkey),
		MaxUploadMb: maxUploadMb,
	}
	s.templates.ExecuteTemplate(w, "home", data)
}

func (s *server) stats(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Global PipeStats
		Active PipeStats
	}{
		Global: s.allPipes.GlobalStats(),
		Active: s.allPipes.ActiveStats(),
	}
	s.templates.ExecuteTemplate(w, "stats", data)
}

// receive data from any senders
func (s *server) recv(w http.ResponseWriter, r *http.Request, p *params) {
	// this is required so that data is streamed back to the client
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// this is used to flush output back to the client as it is received
	flusher, _ := w.(http.Flusher)

	// store the active streams by key so that data can be sent by another request
	receiver := MakeReceiver(w, flusher, p.id, p.interactive, p.username)
	pipe := s.allPipes.AddReceiver(p.key, receiver)
	defer s.allPipes.RemoveReceiver(p.key, receiver)

	// in failure mode, don't allow a connection if there are no senders
	if p.failure && pipe.SenderCount() < 1 {
		http.Error(w, "No senders connected", http.StatusInternalServerError)
		return
	}

	select {
	// the receiver disconnected before completion
	case <-r.Context().Done():
	// a sender completed a transfer and closed the stream (EOF received)
	case <-receiver.CloseNotify():
	}
}

// send data to any connected receivers
func (s *server) send(w http.ResponseWriter, r *http.Request, p *params) {
	// Look to see if there are any receivers attached to this key
	pipe := s.allPipes.AddSender(p.key)
	defer s.allPipes.RemoveSender(p.key, pipe)

	// in failure mode, don't allow a connection if there are no recievers
	if p.failure && pipe.ReceiverCount() < 1 {
		http.Error(w, "No receivers connected", http.StatusExpectationFailed)
		return
	}

	// in block mode, wait for a receiver to connect
	if p.block && pipe.ReceiverCount() < 1 {
		receiverAdded := pipe.ReceiverAddedSubscribe()
		defer pipe.ReceiverAddedUnSubscribe(receiverAdded)
		select {
		// the receiver disconnected before completion
		case <-r.Context().Done():
			return
		// allow a timeout if the receiver disconnected without closing the context
		case <-time.After(24 * time.Hour):
			return
		// a sender was added to the pipe - continue on
		case <-receiverAdded:
		}
	}

	// upload size limit
	body := http.MaxBytesReader(w, r.Body, maxUploadMb*1024*1024)

	// copy the request body to all senders
	sender := MakeSender(pipe, p.id, p.username)
	go sender.Copy(body)

	// The 100-continue message is sent on the first read from the Copy goroutine above
	// A short delay is needed to ensure that it goes out before any data is writen back
	time.Sleep(10 * time.Millisecond)

	s.recv(w, r, p)
}

func main() {
	// Accept a command line flag "-httpaddr :8080"
	// This flag tells the server the http address to listen on
	httpaddr := flag.String("httpaddr", "localhost:8080",
		"the address/port to listen on for http \n"+
			"use :<port> to listen on all addresses\n")

	// Accept a command line flag "-baseurl https://mysite.com/"
	baseurl := flag.String("baseurl", "http://localhost:8080/",
		"the base url of the service \n")

	flag.Parse()

	s := server{
		allPipes:  MakePipeCollection(),
		baseURL:   *baseurl,
		maxID:     0,
		templates: templates(),
	}
	http.HandleFunc("/stats", s.stats)
	http.HandleFunc("/", s.handler)

	log.Println("Listening on http:", *httpaddr)
	log.Fatal(http.ListenAndServe(*httpaddr, nil))
}
