package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Println("usage: client https://pipeto.me/<code>")
		return
	}

	url := os.Args[1]
	fmt.Printf("connected to: %s\n", url)

	pr, pw := io.Pipe()
	go io.Copy(pw, os.Stdin)
	req, _ := http.NewRequest(http.MethodPut, url, ioutil.NopCloser(pr))
	resp, _ := http.DefaultClient.Do(req)
	io.Copy(os.Stdout, resp.Body)
}
