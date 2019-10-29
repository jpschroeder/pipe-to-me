package main

import (
	"io"
)

// Hold the information for a single sender

type Sender struct {
	id   int
	pipe *Pipe
}

// write the buffer to all registered receivers
func (s Sender) Write(buffer []byte) (int, error) {
	for receiver := range s.pipe.receivers {
		// don't send the message to yourself
		if receiver.Id() != s.id {
			// errors from one of the receivers shouldn't affect any others
			receiver.Write(buffer)
		}
	}
	bytes := len(buffer)
	s.pipe.bytes += bytes
	s.pipe.written.WriteCompleted(bytes)
	return bytes, nil
}

// close all of the registered receivers
func (s Sender) Close() error {
	for receiver := range s.pipe.receivers {
		// errors from one of the receivers shouldn't affect any others
		receiver.Close()
	}
	return nil
}

func (s Sender) Copy(reader io.Reader) {
	// copy the body to any listening receivers (see Receivers.Write)
	_, err := io.Copy(s, reader)

	// if the copy made it all the way to EOF, close the receivers
	if err == nil {
		s.Close()
	}
}

func MakeSender(p *Pipe, id int) Sender {
	return Sender{
		pipe: p,
		id:   id,
	}
}