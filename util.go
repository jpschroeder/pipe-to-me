package main

import (
	"crypto/rand"
	"fmt"
)

const keyBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Utilities

func randKey(n int) []byte {
	output := make([]byte, n)
	// We will take n bytes, one byte for each character of output.
	randomness := make([]byte, n)
	// read all random
	_, err := rand.Read(randomness)
	if err != nil {
		panic(err)
	}
	l := len(keyBytes)
	// fill output
	for pos := range output {
		// get random item
		random := uint8(randomness[pos])
		// random % 64
		randomPos := random % uint8(l)
		// put into output
		output[pos] = keyBytes[randomPos]
	}
	return output
}

func getUsername(username string, id int) string {
	if len(username) == 0 {
		return fmt.Sprintf("client %d", id)
	}
	return username
}
