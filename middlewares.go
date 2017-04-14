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

	"github.com/yaacov/mohawk/gziphandler"
)

// GZipper middleware that will gzip http requests
type GZipper struct {
	next http.Handler
}

// SetNext set next http serve func
func (g *GZipper) SetNext(h http.Handler) {
	g.next = h
}

// ServeHTTP http serve func
func (g *GZipper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gziphandler.New(g.next.ServeHTTP)(w, r)
}

// Logger middleware that will log http requests
type Logger struct {
	next http.Handler
}

// SetNext set next http serve func
func (l *Logger) SetNext(h http.Handler) {
	l.next = h
}

// ServeHTTP http serve func
func (l Logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %4s  Accept-Encoding: %s, %s", r.RemoteAddr, r.Method, r.Header.Get("Accept-Encoding"), r.URL)
	l.next.ServeHTTP(w, r)
}

// BadRequest will be called if no route found
type BadRequest struct{}

// ServeHTTP http serve func
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
