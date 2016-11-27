package router

import (
	"log"
	"net/http"
	"net/url"
	"strings"
)

type route struct {
	method   string
	segments []string
	handler  func(http.ResponseWriter, *http.Request, map[string]string)
}

type Router struct {
	Prefix           string
	HandleBadRequest func(http.ResponseWriter, *http.Request, map[string]string)
	Routes           []route
}

func (r *Router) Add(method string, path string, handler func(http.ResponseWriter, *http.Request, map[string]string)) {
	r.Routes = append(r.Routes, route{method, strings.Split(path, "/"), handler})
}

func (r Router) Match(route route, method string, segments []string) (bool, map[string]string) {
	argv := map[string]string{}

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

func (r Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	log.Printf("%s %4s %s", request.RemoteAddr, request.Method, request.URL)

	// get the path
	path := request.URL.EscapedPath()
	argv := map[string]string{}

	// check path prefix
	if !strings.HasPrefix(path, r.Prefix) {
		r.HandleBadRequest(writer, request, argv)
		return
	}

	// clean the path
	path = path[len(r.Prefix):]
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	// try to match the path to a route
	segments := strings.Split(path, "/")
	for _, route := range r.Routes {
		found, argv := r.Match(route, request.Method, segments)

		// if found a match, run the handler for this route
		if found {
			route.handler(writer, request, argv)
			return
		}
	}

	// handle page not found
	r.HandleBadRequest(writer, request, argv)
}
