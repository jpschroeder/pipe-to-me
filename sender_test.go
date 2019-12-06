package main

import (
	"testing"
)

func TestSenderWrite(t *testing.T) {
	handler := &TestHandler{}
	pipe := MakePipe(handler)
	receivers := [3]*TestReceiver{
		&TestReceiver{},
		&TestReceiver{},
		&TestReceiver{},
	}
	for _, r := range receivers {
		pipe.AddReceiver(r)
	}
	sender := MakeSender(pipe, 1, "")

	input := "test input 1"
	count, err := sender.Write([]byte(input))

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

	sender.Write([]byte(input))
	if pipe.BytesSent() != len(input)*2 {
		t.Errorf("Invalid pipe bytecount: %d %d", len(input)*2, pipe.BytesSent())
	}
}

func TestSenderClose(t *testing.T) {
	handler := &TestHandler{}
	pipe := MakePipe(handler)
	receivers := [3]*TestReceiver{
		&TestReceiver{},
		&TestReceiver{},
		&TestReceiver{},
	}
	for _, r := range receivers {
		pipe.AddReceiver(r)
	}
	sender := MakeSender(pipe, 1, "")

	sender.Write([]byte("test input"))
	err := sender.Close()

	if err != nil {
		t.Errorf("Error closing pipe: %s", err.Error())
	}

	for _, r := range receivers {
		if r.closeCount != 1 {
			t.Errorf("Pipe close not called appropriately %d %d", 1, r.closeCount)
		}
	}
}
