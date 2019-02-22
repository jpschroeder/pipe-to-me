package main

import (
	"fmt"
	"io"
	"net/http"
)

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `# pipe to cloud
something | curl -fsS -T - http://%s/pipe

# pipe from cloud
curl -s http://%s/pipe`, r.Host, r.Host)
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

func send(w http.ResponseWriter, r *http.Request) {
	//r.Body = http.MaxBytesReader(w, r.Body, 5000000) // 10 megabytes
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

	done := false
	for done == false {
		select {
		case <-w.(http.CloseNotifier).CloseNotify():
			done = true
			break
		case <-receiver.done:
			done = true
			break
		}
	}
	delete(receivers, receiver)
	fmt.Println("Receiver Disconnected: Notify")
}

func main() {
	receivers = make(map[*Receiver]bool)
	http.HandleFunc("/pipe", pipe)
	http.HandleFunc("/", home)
	http.ListenAndServe(":1313", nil)
}
