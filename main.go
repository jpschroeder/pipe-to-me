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

type Receiver struct {
	writer  io.Writer
	flusher http.Flusher
	done    chan bool
}

var receivers map[*Receiver]bool

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
	receiver := &Receiver{
		writer:  w,
		flusher: flusher,
		done:    make(chan bool),
	}

	receivers[receiver] = true
	defer delete(receivers, receiver)

	done := false
	for done == false {
		select {
		// The reciever disconnected themselves
		case <-w.(http.CloseNotifier).CloseNotify():
			done = true
			break
		// EOF was received on one of the send channels
		case <-receiver.done:
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
	copy(receivers, input)
	fmt.Println("Sender Disconnected")
}

func copy(dst map[*Receiver]bool, src io.Reader) {
	buf := make([]byte, 32*1024)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			fmt.Println("Sender Data Received: ", nr)
			for receiver := range dst {
				nw, _ := receiver.writer.Write(buf[0:nr])
				if nw > 0 {
					receiver.flusher.Flush()
				}
			}
		}
		if er == io.EOF {
			for receiver := range dst {
				receiver.flusher.Flush()
				receiver.done <- true
			}
		}
		if er != nil {
			break
		}
	}
}

func main() {
	receivers = make(map[*Receiver]bool)
	http.HandleFunc("/pipe", pipe)
	http.HandleFunc("/", home)
	http.ListenAndServe(":1313", nil)
}
