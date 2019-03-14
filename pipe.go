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
	written   WriteCompleteHandler
	// a list of channels that want to be notified of new receivers
	receiverAdded map[chan bool]bool
}

// add a new receiver listening on the pipe
func (p *Pipe) AddReceiver(w io.WriteCloser) {
	p.receivers[w] = true
	p.ReceiverAddedNotify()
}

// remove a previously added receiver
func (p *Pipe) RemoveReceiver(w io.WriteCloser) {
	delete(p.receivers, w)
}

// the number of receivers on the pipe
func (p Pipe) ReceiverCount() int {
	return len(p.receivers)
}

// listen for new receivers
func (p *Pipe) ReceiverAddedSubscribe() chan bool {
	channel := make(chan bool)
	p.receiverAdded[channel] = true
	return channel
}

// stop listening for new receivers
func (p *Pipe) ReceiverAddedUnSubscribe(channel chan bool) {
	delete(p.receiverAdded, channel)
}

// notify all listeners that a receiver was added
func (p *Pipe) ReceiverAddedNotify() {
	for channel := range p.receiverAdded {
		// non-blocking
		select {
		case channel <- true:
		default:
		}
	}
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
	bytes := len(buffer)
	p.bytes += bytes
	p.written.WriteCompleted(bytes)
	return bytes, nil
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

func MakePipe(written WriteCompleteHandler) *Pipe {
	return &Pipe{
		receivers:     make(map[io.WriteCloser]bool),
		senders:       0,
		bytes:         0,
		written:       written,
		receiverAdded: make(map[chan bool]bool),
	}
}
