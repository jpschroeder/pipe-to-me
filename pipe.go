package main

import (
	"fmt"
)

// Pipe holds the information for a single pipe
type Pipe struct {
	// a list of receivers that are listening on a pipe
	// allow receivers to be added an removed dynamically
	receivers map[RecieveWriter]bool
	senders   int
	bytes     int
	written   WriteCompleteHandler
	// a list of channels that want to be notified of new receivers
	receiverAdded map[chan bool]bool
}

// AddReceiver adds a new receiver listening on the pipe
func (p *Pipe) AddReceiver(w RecieveWriter) {
	p.receivers[w] = true
	p.Write(Message{
		fromID: w.ID(),
		fromUser: w.Username(),
		buffer: []byte("connected\n"),
		system: true })
	p.ReceiverAddedNotify()
}

// RemoveReceiver removes a previously added receiver
func (p *Pipe) RemoveReceiver(w RecieveWriter) {
	delete(p.receivers, w)
	p.Write(Message{
		fromID: w.ID(),
		fromUser: w.Username(),
		buffer: []byte("disconnected\n"),
		system: true })
}

// ReceiverCount returns the number of receivers on the pipe
func (p Pipe) ReceiverCount() int {
	return len(p.receivers)
}

// ReceiverAddedSubscribe listens for new receivers
func (p *Pipe) ReceiverAddedSubscribe() chan bool {
	channel := make(chan bool)
	p.receiverAdded[channel] = true
	return channel
}

// ReceiverAddedUnSubscribe stops listening for new receivers
func (p *Pipe) ReceiverAddedUnSubscribe(channel chan bool) {
	delete(p.receiverAdded, channel)
}

// ReceiverAddedNotify notifies all listeners that a receiver was added
func (p *Pipe) ReceiverAddedNotify() {
	for channel := range p.receiverAdded {
		// non-blocking
		select {
		case channel <- true:
		default:
		}
	}
}

// AddSender adds a new sender connected to send data on the pipe (informational)
func (p *Pipe) AddSender() {
	p.senders++
}

// RemoveSender removes a sender connected to the pipe (informational)
func (p *Pipe) RemoveSender() {
	p.senders--
}

// SenderCount returns the number of senders on the pipe
func (p Pipe) SenderCount() int {
	return p.senders
}

// BytesSent returns the number of bytes sent through the pipe
func (p Pipe) BytesSent() int {
	return p.bytes
}

// Write the buffer to all registered receivers
func (p *Pipe) Write(m Message) (int, error) {
	for receiver := range p.receivers {
		receiver.Write(m.Format(receiver))
	}
	bytes := len(m.buffer)
	p.bytes += bytes
	p.written.WriteCompleted(bytes)
	return bytes, nil
}

// Close all of the registered receivers
func (p *Pipe) Close() error {
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

// MakePipe creates the struct for a pipe
func MakePipe(written WriteCompleteHandler) *Pipe {
	return &Pipe{
		receivers:     make(map[RecieveWriter]bool),
		senders:       0,
		bytes:         0,
		written:       written,
		receiverAdded: make(map[chan bool]bool),
	}
}
