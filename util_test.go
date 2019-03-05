package main

import (
	"bytes"
	"testing"
)

func TestTest(t *testing.T) {
	//t.Errorf("test test")
}

func TestRandKey(t *testing.T) {
	key := randKey(keySize)
	if len(key) != keySize {
		t.Errorf("Invalid key size %d - %d", keySize, len(key))
	}

	anotherkey := randKey(keySize)
	if bytes.Compare(key, anotherkey) == 0 {
		t.Errorf("Duplicate key generated")
	}
}
