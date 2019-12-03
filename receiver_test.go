package main

import (
	"bytes"
	"testing"
)

type TestFlusher struct {
	flushCount int
}

func (t *TestFlusher) Flush() {
	t.flushCount++
}

func TestReceiverWrite(t *testing.T) {
	var w bytes.Buffer
	f := TestFlusher{flushCount: 0}
	receiver := MakeReceiver(&w, &f, 0, false)

	input := "test input"
	count, err := receiver.Write([]byte(input))

	if err != nil {
		t.Errorf("Error writing to receiver: %s", err.Error())
	}

	if count != len(input) {
		t.Errorf("Invalid length written to receiver: %d %d", len(input), count)
	}

	if w.String() != input {
		t.Errorf("Invalid string written to receiver: %s %s", input, w.String())
	}

	if f.flushCount != 1 {
		t.Errorf("Flush not called: %d", f.flushCount)
	}
}

func TestReceiverClose(t *testing.T) {
	var w bytes.Buffer
	f := TestFlusher{flushCount: 0}
	receiver := MakeReceiver(&w, &f, 0, false)

	go func() {
		receiver.Write([]byte("test input"))
		receiver.Close()
	}()

	closed := <-receiver.CloseNotify()

	if !closed {
		t.Errorf("Receiver was not closed")
	}

	if f.flushCount != 2 {
		t.Errorf("Final flush not called appropriately: %d", f.flushCount)
	}
}
