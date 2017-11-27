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

// Package middleware
package middleware

import (
	"net/http"
)

// HandlerFunc implements ServeHTTP and it's actually a function that its
// signature is similar to http.HandlerFunc
type HandlerFunc func(w http.ResponseWriter, r *http.Request)

func (hf HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hf(w, r)
}

type MiddleWare interface {
	SetNext(http.Handler)
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// append concat a list of MiddleWares into the router routing table
func Append(handlers []MiddleWare) {
	switch len(handlers) {
	case 0:
		fallthrough
	case 1:
		return
	default:
		handlers[0].SetNext(handlers[1])
		Append(handlers[1:])
	}
}
