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

// Package middleware middlewares for MoHawk
package middleware

import (
	"log"
	"net/http"
)

// Logger middleware that will log http requests
type Logger struct {
	Verbose bool
	next    http.Handler
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
