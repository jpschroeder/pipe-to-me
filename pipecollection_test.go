package main

import "testing"

func TestPipeCollectionWrite(t *testing.T) {
	pipes := MakePipeCollection()
	receiver := &TestReceiver{}
	pipes.AddReceiver("key", receiver)

	pipe := pipes.AddSender("key")
	sender := MakeSender(pipe, 1)
	input := "test input"
	sender.Write([]byte(input))

	if receiver.writer.String() != input {
		t.Errorf("Invalid collection write: %s %s", input, receiver.writer.String())
	}
}

func TestMultipleReadersAndWriters(t *testing.T) {
	input := "test input"
	pipes := MakePipeCollection()

	p1 := pipes.AddSender("key1")
	s1 := MakeSender(p1, 1)
	r1 := &TestReceiver{}
	pipes.AddReceiver("key1", r1)

	p2 := pipes.AddSender("key1")
	s2 := MakeSender(p2, 1)
	r2 := &TestReceiver{}
	pipes.AddReceiver("key1", r2)

	s1.Write([]byte(input))
	s2.Write([]byte(input))

	if r1.writer.String() != (input + input) {
		t.Errorf("Invalid collection write (r1): %s %s", input+input, r1.writer.String())
	}
	if r2.writer.String() != (input + input) {
		t.Errorf("Invalid collection write (r2): %s %s", input+input, r2.writer.String())
	}
}

func TestCollectionAddRemove(t *testing.T) {
	pipes := MakePipeCollection()

	pipes.AddSender("key1")

	r1 := &TestReceiver{}
	pipes.AddReceiver("key1", r1)
	pipes.RemoveReceiver("key1", r1)

	p2 := pipes.AddSender("key2")
	pipes.RemoveSender("key2", p2)

	r2 := &TestReceiver{}
	pipes.AddReceiver("key2", r2)

	p3 := pipes.AddSender("key2")
	pipes.RemoveSender("key2", p3)

	r3 := &TestReceiver{}
	pipes.AddReceiver("key2", r3)
	pipes.RemoveReceiver("key2", r3)

	stats := pipes.ActiveStats()

	if stats.PipeCount != 2 {
		t.Errorf("Invalid pipe count: %d", stats.PipeCount)
	}

	if stats.SenderCount != 1 {
		t.Errorf("Invalid sender count: %d", stats.SenderCount)
	}

	if stats.ReceiverCount != 1 {
		t.Errorf("Invalid receiver count: %d", stats.ReceiverCount)
	}
}
