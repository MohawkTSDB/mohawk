// Copyright 2016 Red Hat, Inc. and/or its affiliates
// and other contributors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package main
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Logger middleware that will log http requests
type Logger struct {
	next http.Handler
}

func (l *Logger) SetNext(h http.Handler) {
	l.next = h
}

func (l Logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %4s %s", r.RemoteAddr, r.Method, r.URL)
	l.next.ServeHTTP(w, r)
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
