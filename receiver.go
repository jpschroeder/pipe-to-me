package main

import (
	"io"
	"net/http"
)

// RecieveWriter is an interface that allows writing to a receiver
// it is implemented by Receiver
type RecieveWriter interface {
	io.WriteCloser
	ID() int
	Interactive() bool
	Username() string
}

// Receiver holds the information for a single receiver
// a writer that is automatically flushed back to the receiver client
// and a notification channel when it is closed
type Receiver struct {
	id          int
	interactive bool
	username    string
	writer      io.Writer
	flusher     http.Flusher
	done        chan bool
}

// ID returns the identifier for this reader
func (r Receiver) ID() int {
	return r.id
}

// Interactive returns whether or not to show connect/disconnect messages to the receiver
func (r Receiver) Interactive() bool {
	return r.interactive
}

// Username returns the username supplied by the receiver (or client <id> if none was supplied)
func (r Receiver) Username() string {
	return getUsername(r.username, r.id)
}

// Write a single received buffer to the receiver and flush it back to the client
func (r Receiver) Write(p []byte) (n int, err error) {
	if len(p) < 1 {
		return
	}
	n, err = r.writer.Write(p)
	r.flusher.Flush()
	return
}

// Close the receiver. flush it one last time and notify that it is closed
func (r Receiver) Close() error {
	r.flusher.Flush()
	r.done <- true
	return nil
}

// CloseNotify returns a notification channel that will tell when the reciever has been closed
func (r Receiver) CloseNotify() <-chan bool {
	return r.done
}

// MakeReceiver creates a new receiver struct
func MakeReceiver(w io.Writer, f http.Flusher, id int, interactive bool, username string) Receiver {
	return Receiver{
		writer:      w,
		flusher:     f,
		id:          id,
		interactive: interactive,
		username:    username,
		done:        make(chan bool),
	}
}
