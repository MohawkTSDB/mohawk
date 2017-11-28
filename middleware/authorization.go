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

func AuthDecorator(token, publicPathRegex string) Decorator {
	publicPath := regexp.MustCompile(publicPathRegex)
	authToken := "Bearer " + token
	return Decorator(func(h http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println(r.Header.Get("Authorization"), authToken, r.Header.Get("Authorization") == authToken)
			if r.Header.Get("Authorization") == authToken || publicPath.MatchString(r.URL.EscapedPath()) {
				h(w, r)
			}

			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "Unauthorized - 401\n")
		})
	})
}
