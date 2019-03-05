package main

import (
	"bytes"
	"testing"
)

func TestPipeSenders(t *testing.T) {
	pipe := MakePipe()
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

type TestReceiver struct {
	writer     bytes.Buffer
	closeCount int
}

func (r *TestReceiver) Write(p []byte) (n int, err error) {
	n, err = r.writer.Write(p)
	return
}

func (r *TestReceiver) Close() error {
	r.closeCount++
	return nil
}

func TestPipeReceivers(t *testing.T) {
	pipe := MakePipe()
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

func TestPipeWrite(t *testing.T) {
	pipe := MakePipe()
	receivers := [3]*TestReceiver{
		&TestReceiver{},
		&TestReceiver{},
		&TestReceiver{},
	}
	for _, r := range receivers {
		pipe.AddReceiver(r)
	}

	input := "test input 1"
	count, err := pipe.Write([]byte(input))

	if err != nil {
		t.Errorf("Error writing to pipe: %s", err.Error())
	}

	if count != len(input) {
		t.Errorf("Invalid length written to pipe: %d %d", len(input), count)
	}

	for _, r := range receivers {
		if r.writer.String() != input {
			t.Errorf("Invalid string written to receiver: %s %s", input, r.writer.String())
		}
	}

	pipe.Write([]byte(input))
	if pipe.BytesSent() != len(input)*2 {
		t.Errorf("Invalid pipe bytecount: %d %d", len(input)*2, pipe.BytesSent())
	}
}

func TestPipeClose(t *testing.T) {
	pipe := MakePipe()
	receivers := [3]*TestReceiver{
		&TestReceiver{},
		&TestReceiver{},
		&TestReceiver{},
	}
	for _, r := range receivers {
		pipe.AddReceiver(r)
	}

	pipe.Write([]byte("test input"))
	err := pipe.Close()

	if err != nil {
		t.Errorf("Error closing pipe: %s", err.Error())
	}

	for _, r := range receivers {
		if r.closeCount != 1 {
			t.Errorf("Pipe close not called appropriately %d %d", 1, r.closeCount)
		}
	}
}
