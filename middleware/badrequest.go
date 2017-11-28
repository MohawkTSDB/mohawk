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

// Package middleware middlewares for Mohawk
package middleware

import (
	"fmt"
	"net/http"
)

// BadRequestHandler handler
func BadRequestHandler(logFunc func(string, ...interface{})) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// we return 200 for any OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(200)
			return
		}
		logFunc("Page not found - 404:\n")
		w.WriteHeader(404)
		fmt.Fprintf(w, "Page not found - 404\n")
	}
}
