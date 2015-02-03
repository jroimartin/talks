package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jroimartin/orujo"
)

// START OMIT
func main() {
	s := orujo.NewServer("localhost:8080")

	s.Route(`^/$`, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello world!")
	}))

	log.Fatalln(s.ListenAndServe())
}

// STOP OMIT
