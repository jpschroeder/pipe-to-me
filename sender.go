package main

import (
	"io"
)

// Sender holds the information for a single sender
type Sender struct {
	id   int
	pipe *Pipe
}

// Write the buffer to all registered receivers
func (s Sender) Write(buffer []byte) (int, error) {
	return s.pipe.Write(buffer, s.id, false)
}

// Close all of the registered receivers
func (s Sender) Close() error {
	return s.pipe.Close()
}

// Copy transfers bytes from the reader to the attached pipe
func (s Sender) Copy(reader io.Reader) {
	// copy the body to any listening receivers (see Receivers.Write)
	_, err := io.Copy(s, reader)

	// if the copy made it all the way to EOF, close the receivers
	if err == nil {
		s.Close()
	}
}

// MakeSender creates a new sender
func MakeSender(p *Pipe, id int) Sender {
	return Sender{
		pipe: p,
		id:   id,
	}
}
