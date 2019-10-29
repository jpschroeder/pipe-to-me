package main

import (
	"bytes"
	"testing"
)

type TestReceiver struct {
	writer     bytes.Buffer
	closeCount int
}

func (r TestReceiver) Id() int {
	return 0
}

func (r *TestReceiver) Write(p []byte) (n int, err error) {
	n, err = r.writer.Write(p)
	return
}

func (r *TestReceiver) Close() error {
	r.closeCount++
	return nil
}

type TestHandler struct {
	bytes int
}

func (pc *TestHandler) WriteCompleted(bytes int) {
	pc.bytes += bytes
}

func TestPipeSenders(t *testing.T) {
	handler := &TestHandler{}
	pipe := MakePipe(handler)
	pipe.AddSender()
	if pipe.SenderCount() != 1 {
		t.Errorf("Invalid sender count: %d %d", 1, pipe.SenderCount())
	}

	pipe.AddSender()
	pipe.AddSender()
	pipe.AddSender()

	if pipe.SenderCount() != 4 {
		t.Errorf("Invalid sender count: %d %d", 4, pipe.SenderCount())
	}

	pipe.RemoveSender()
	pipe.RemoveSender()
	pipe.AddSender()
	pipe.RemoveSender()

	if pipe.SenderCount() != 2 {
		t.Errorf("Invalid sender count: %d %d", 2, pipe.SenderCount())
	}
}

func TestPipeReceivers(t *testing.T) {
	handler := &TestHandler{}
	pipe := MakePipe(handler)
	r1 := &TestReceiver{}
	r2 := &TestReceiver{}
	r3 := &TestReceiver{}

	pipe.AddReceiver(r1)
	if pipe.ReceiverCount() != 1 {
		t.Errorf("Invalid receiver count: %d %d", 1, pipe.ReceiverCount())
	}

	pipe.AddReceiver(r2)
	pipe.AddReceiver(r3)
	if pipe.ReceiverCount() != 3 {
		t.Errorf("Invalid receiver count: %d %d", 3, pipe.ReceiverCount())
	}

	pipe.RemoveReceiver(r1)
	pipe.RemoveReceiver(r2)
	if pipe.ReceiverCount() != 1 {
		t.Errorf("Invalid receiver count: %d %d", 1, pipe.ReceiverCount())
	}

	pipe.AddReceiver(r3) // add the same receiver
	if pipe.ReceiverCount() != 1 {
		t.Errorf("Invalid receiver count: %d %d", 1, pipe.ReceiverCount())
	}
}
