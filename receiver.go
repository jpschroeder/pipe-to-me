package main

import (
	"io"
	"net/http"
)

type RecieveWriter interface {
	io.WriteCloser
	Id() int
}

// Hold the information for a single receiver

// a writer that is automatically flushed back to the receiver client
// and a notification channel when it is closed
type Receiver struct {
	id      int
	writer  io.Writer
	flusher http.Flusher
	done    chan bool
}

func (r Receiver) Id() int {
	return r.id
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

func MakeReceiver(w io.Writer, f http.Flusher, id int) Receiver {
	return Receiver{
		writer:  w,
		flusher: f,
		id:      id,
		done:    make(chan bool),
	}
}
