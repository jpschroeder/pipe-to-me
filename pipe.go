package main

import (
	"fmt"
	"io"
)

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
