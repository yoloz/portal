package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"
)

func TestStripPrefix(t *testing.T) {
	res, err := http.Get("http://localhost:10010/note/README.md")
	if err != nil {
		log.Fatal(err)
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", body)
}
