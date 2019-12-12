package main

// Message contains all fields necessary to render a message
type Message struct {
	fromID   int
	fromUser string
	buffer   []byte
	system   bool
}

// Format customizes the message for a particular receiver
func (m Message) Format(receiver RecieveWriter) []byte {
	if receiver.Interactive() {
		return m.formatInteractive(receiver)
	}
	return m.formatNonInteractive(receiver)
}

func (m Message) formatInteractive(receiver RecieveWriter) []byte {
	// Echo back system messages only
	if m.fromID == receiver.ID() && !m.system {
		return []byte{}
	}
	// Add username
	if len(m.fromUser) > 0 {
		m.buffer = append([]byte(m.fromUser+": "), m.buffer...)
	}
	return m.buffer
}

func (m Message) formatNonInteractive(receiver RecieveWriter) []byte {
	// Don't send system messages
	if m.system {
		return []byte{}
	}
	// Don't echo messages back to the sender
	if m.fromID == receiver.ID() {
		return []byte{}
	}
	return m.buffer
}
