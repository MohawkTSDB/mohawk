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
	"net/http"
)

// Handler handler type interface
type Handler interface {
	SetNext(http.Handler)
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// Append concat a list of Routers into the router routing table
// returns
// 	http.HandlerFunc - the first http handler function to call
func Append(handlers ...Handler) http.HandlerFunc {
	listSize := len(handlers)

	// concat all routes, last item has no next function
	for ix, r := range handlers[:listSize-1] {
		r.SetNext(handlers[ix+1])
	}

	return handlers[0].ServeHTTP
}
