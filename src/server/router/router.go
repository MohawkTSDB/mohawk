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

// Package router for http request routing
package router

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/MohawkTSDB/mohawk/src/api_errors"
)

type route struct {
	method   string
	segments []string
	handler  func(http.ResponseWriter, *http.Request, map[string]string) error
}

// Router ah http request router
type Router struct {
	Prefix string
	Routes []route
	next   http.Handler
}

// Add add a new route into the router routing table
func (r *Router) Add(method string, path string, handler func(http.ResponseWriter, *http.Request, map[string]string) error) {
	r.Routes = append(r.Routes, route{method, strings.Split(path, "/"), handler})
}

func handleError(err error, w http.ResponseWriter) {
	if mErr, ok := err.(apiErrors.Error); ok {
		mErr.JSON(w)
		return
	}
	w.WriteHeader(500)
	w.Write([]byte(http.StatusText(500)))
}

// match match a request to a route, and parse the arguments embedded in the route path
// returns
// 	bool - true if route match
// 	map  - a map of arguments parsed from the route path
func (r Router) match(route route, method string, segments []string) (bool, map[string]string) {
	argv := make(map[string]string)

	// check method and segments list length
	if method != route.method || len(segments) != len(route.segments) {
		return false, argv
	}

	// check segments
	for i, segment := range route.segments {
		if segment[0] == ':' {
			// if this is an argument segments, parse it
			value, _ := url.QueryUnescape(segments[i])
			argv[segment[1:]] = value
		} else if segments[i] != segment {
			// if this segment does not match the route exit
			return false, argv
		}
	}

	// found matching route
	return true, argv
}

// SetNext sets the next handler in the routing list
func (r *Router) SetNext(h http.Handler) {
	r.next = h
}

// ServeHTTP try to match a route to the request, and call the route handler
func (r Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// get the path
	path := request.URL.EscapedPath()

	// check path prefix
	if !strings.HasPrefix(path, r.Prefix) {
		r.next.ServeHTTP(writer, request)
		return
	}

	// clean the path
	path = path[len(r.Prefix):]
	if len(path) > 0 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	// try to match the path to a route
	segments := strings.Split(path, "/")
	for _, route := range r.Routes {
		found, argv := r.match(route, request.Method, segments)

		// if found a match, run the handler for this route
		if found {
			if err := route.handler(writer, request, argv); err != nil {
				handleError(err, writer)
			}
			return
		}
	}

	// handle page not found
	r.next.ServeHTTP(writer, request)
}
