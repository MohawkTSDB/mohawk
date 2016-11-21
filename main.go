package main

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/yaacov/mohawk/backends"
)

const VER = "0.21.0"

type Router struct {
	prefix  string
	backend backend.Backend
}

func (h Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %4s %s", r.RemoteAddr, r.Method, r.URL)

	// parse request
	// -------------

	// sanity check
	path := r.URL.EscapedPath()
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	if !strings.HasPrefix(path, h.prefix) {
		h.handleBadRequest(w, r)
		return
	}

	// parse
	r.ParseForm()
	argv := strings.Split(path[len(h.prefix):], "/")
	argc := len(argv)
	for _, e := range argv {
		arg, _ := url.QueryUnescape(e)
		r.Form.Add("_argv", arg)
	}

	// route
	//------

	// get metrics list
	if r.Method == "GET" && argc == 1 && argv[0] == "status" {
		h.handleStatus(w, r)
		return
	}

	// get metrics list
	if r.Method == "GET" && argc == 1 && argv[0] == "metrics" {
		h.handleList(w, r)
		return
	}

	// get data items
	if r.Method == "GET" && argc == 3 &&
		(argv[0] == "gauges" || argv[0] == "strings" || argv[0] == "availability") {
		h.handleGetData(w, r)
		return
	}

	// push data
	if r.Method == "POST" && argc == 2 &&
		(argv[0] == "gauges" || argv[0] == "strings" || argv[0] == "availability") {
		h.handlePushData(w, r)
		return
	}

	// push tags
	if r.Method == "PUT" && argc == 3 &&
		(argv[0] == "gauges" || argv[0] == "strings" || argv[0] == "availability") {
		h.handleUpdateTags(w, r)
		return
	}

	// handle page not found
	h.handleBadRequest(w, r)
}

func main() {
	b := backend.Random{}
	b.Open()

	srv := &http.Server{
		Addr:           ":8443",
		Handler:        Router{prefix: "/hawkular/metrics/", backend: b},
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := srv.ListenAndServeTLS("server.pem", "server.key")
	log.Fatal(err)
}
