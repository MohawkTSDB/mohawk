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
	"strings"

	"github.com/yaacov/mohawk/middleware/gziphandler"
)

// GZipper middleware that will gzip http requests
type GZipper struct {
	Verbose bool
	next    http.Handler
}

// SetNext set next http serve func
func (g *GZipper) SetNext(h http.Handler) {
	g.next = h
}

// ServeHTTP http serve func
func (g *GZipper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		g.next.ServeHTTP(w, r)
		return
	}

	if g.Verbose {
		log.Printf("Using gzip encoding")
	}

	gziphandler.New(g.next.ServeHTTP)(w, r)
}
