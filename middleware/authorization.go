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

// Package middleware middlewares for Mohawk
package middleware

import (
	"fmt"
	"net/http"
	"regexp"
)

// Authorization middleware that will authorize http requests
type Authorization struct {
	UseToken        bool
	Verbose         bool
	PublicPathRegex string
	Token           string
	next            http.Handler
}

// SetNext set next http serve func
func (a *Authorization) SetNext(h http.Handler) {
	a.next = h
}

// ServeHTTP http serve func
func (a *Authorization) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// check authorization
	if !a.UseToken || r.Header.Get("Authorization") == "Bearer "+a.Token {
		a.next.ServeHTTP(w, r)
		return
	}

	// check for public path
	path := r.URL.EscapedPath()
	validRegex := regexp.MustCompile(a.PublicPathRegex)
	if validRegex.MatchString(path) {
		a.next.ServeHTTP(w, r)
		return
	}

	// this is an none uthorized request
	w.WriteHeader(401)
	fmt.Fprintf(w, "Unauthorized - 401\n")
}
