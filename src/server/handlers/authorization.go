// Copyright 2016,2017,2018 Yaacov Zamir <kobi.zamir@gmail.com>
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

// Package handler http server handler functions
package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
)

// Authorization middleware that will authorize http requests
type Authorization struct {
	PublicPathRegex *regexp.Regexp
	Authorization   string
	Verbose         bool
	next            http.Handler
}

// SetNext set next http serve func
func (a *Authorization) SetNext(h http.Handler) {
	a.next = h
}

// ServeHTTP http serve func
func (a *Authorization) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var u interface{}

	if r.Header.Get("Authorization") == a.Authorization || a.PublicPathRegex.MatchString(r.URL.EscapedPath()) {
		a.next.ServeHTTP(w, r)
		return
	}

	log.Printf("Unauthorized - 401\n")

	// printout debug data
	if a.Verbose {
		json.NewDecoder(r.Body).Decode(&u)
		r.ParseForm()

		log.Printf("Request: %+v\n", r)
		log.Printf("Body: %+v\n", u)
	}

	w.WriteHeader(401)
	fmt.Fprintf(w, "{\"error\":\"401\",\"message\":\"Unauthorized - 401\"}")
}
