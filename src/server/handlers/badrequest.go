// Copyright 2016,2017 Yaacov Zamir <kobi.zamir@gmail.com>
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

// Package handler
package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// BadRequest will be called if no route found
type BadRequest struct {
	Verbose bool
}

// SetNext set next http serve func
func (b *BadRequest) SetNext(_h http.Handler) {
}

// ServeHTTP http serve func
func (b BadRequest) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var u interface{}

	log.Printf("Page not found - 404\n")

	// printout debug data
	if b.Verbose {
		json.NewDecoder(r.Body).Decode(&u)
		r.ParseForm()

		log.Printf("Request: %+v\n", r)
		log.Printf("Body: %+v\n", u)
	}

	w.WriteHeader(404)
	fmt.Fprintf(w, "{\"error\":\"404\",\"message\":\"Page not found - 404\"}")
}
