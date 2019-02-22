package main

import (
	"fmt"
	"io"
	"net/http"
)

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
# receiver
curl -s http://%s/pipe

# sender
echo something | curl -T- http://%s/pipe
	`, r.Host, r.Host)
}

// Hold the information for a single receiver
type Receiver struct {
	writer  io.Writer
	flusher http.Flusher
	done    chan bool
}

func (r Receiver) Write(p []byte) (n int, err error) {
	n, err = r.writer.Write(p)
	r.flusher.Flush()
	return
}

func (r Receiver) Close() error {
	r.flusher.Flush()
	r.done <- true
	return nil
}

func (r Receiver) CloseNotify() <-chan bool {
	return r.done
}

func MakeReceiver(w io.Writer, f http.Flusher) Receiver {
	return Receiver{
		writer:  w,
		flusher: f,
		done:    make(chan bool),
	}
}

// Combine multiple receivers together
type Receivers struct {
	writers map[Receiver]bool
}

func (mw *Receivers) Add(w Receiver) {
	mw.writers[w] = true
}

func (mw *Receivers) Delete(w Receiver) {
	delete(mw.writers, w)
}

func (mw Receivers) Write(p []byte) (n int, err error) {
	for writer := range mw.writers {
		n, _ = writer.Write(p)
	}
	return len(p), nil
}

func (mw Receivers) Close() error {
	for writer := range mw.writers {
		writer.Close()
	}
	return nil
}

func MakeReceivers() Receivers {
	allWriters := make(map[Receiver]bool)
	return Receivers{allWriters}
}

// Application Code

var allReceivers Receivers

func pipe(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		recv(w, r)
	} else {
		send(w, r)
	}
}

func recv(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Receiver Connected")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	flusher, _ := w.(http.Flusher)

	receiver := MakeReceiver(w, flusher)
	allReceivers.Add(receiver)
	defer allReceivers.Delete(receiver)

	done := false
	for done == false {
		select {
		// The receiver disconnected themselves
		case <-w.(http.CloseNotifier).CloseNotify():
			done = true
			break
		// EOF was received on one of the send channels
		case <-receiver.CloseNotify():
			done = true
			break
		}
	}
	fmt.Println("Receiver Disconnected")
}

func send(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 5000000) // 5 megabytes
	fmt.Println("Sender Connected")
	input := r.Body
	_, err := io.Copy(allReceivers, input)
	if err == nil {
		allReceivers.Close()
	}
	fmt.Println("Sender Disconnected")
}

func main() {
	allReceivers = MakeReceivers()
	http.HandleFunc("/pipe", pipe)
	http.HandleFunc("/", home)
	http.ListenAndServe(":1313", nil)
}
