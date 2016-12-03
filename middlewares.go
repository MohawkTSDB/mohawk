package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Logger middleware that will log http requests
type Logger struct {
	Next http.Handler
}

func (l Logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %4s %s", r.RemoteAddr, r.Method, r.URL)
	l.Next.ServeHTTP(w, r)
}

// BadRequest will be called if no route found
type BadRequest struct{}

func (b BadRequest) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var u interface{}
	json.NewDecoder(r.Body).Decode(&u)
	r.ParseForm()

	log.Printf("BadRequest:\n")
	log.Printf("Request: %+v\n", r)
	log.Printf("Body: %+v\n", u)

	w.WriteHeader(404)
	fmt.Fprintf(w, "Page not found - 404\n")
}
