package router

import (
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Route struct {
	Method  string
	paths   []string
	handler func(http.ResponseWriter, *http.Request, map[string]string)
}

type Router struct {
	Prefix           string
	HandleBadRequest func(http.ResponseWriter, *http.Request, map[string]string)
	Routes           []Route
}

func (router *Router) Add(method string, path string, handler func(http.ResponseWriter, *http.Request, map[string]string)) {
	router.Routes = append(router.Routes, Route{method, strings.Split(path, "/"), handler})
}

func (_ Router) Match(route Route, paths []string, method string) (bool, map[string]string) {
	argv := map[string]string{}

	// check method
	if method != route.Method || len(paths) != len(route.paths) {
		return false, argv
	}

	// check path
	for i, p := range route.paths {
		if p[0] != ':' && paths[i] != p {
			return false, argv
		}
	}

	// get arguments
	for i, p := range route.paths {
		if p[0] == ':' {
			e, _ := url.QueryUnescape(paths[i])
			argv[p[1:]] = e
		}
	}

	// found matching route
	return true, argv
}

func (router Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %4s %s", r.RemoteAddr, r.Method, r.URL)

	// get the path
	path := r.URL.EscapedPath()
	argv := map[string]string{}

	// check path prefix
	if !strings.HasPrefix(path, router.Prefix) {
		router.HandleBadRequest(w, r, argv)
		return
	}

	// clean the path
	path = path[len(router.Prefix):]
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	// try to match the path to a route
	paths := strings.Split(path, "/")
	for _, route := range router.Routes {
		found, argv := router.Match(route, paths, r.Method)

		// if found a match, run the handler for this route
		if found {
			route.handler(w, r, argv)
			return
		}
	}

	// handle page not found
	router.HandleBadRequest(w, r, argv)
}
