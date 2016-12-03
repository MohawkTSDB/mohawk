package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/yaacov/mohawk/backends"
	"github.com/yaacov/mohawk/router"
)

// VER the server version
const VER = "0.2.0"

// BadRequest will be called if no route found
type BadRequest struct{}

func (b BadRequest) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var u interface{}
	json.NewDecoder(r.Body).Decode(&u)
	r.ParseForm()

	log.Printf("BadRequest:\n")
	log.Printf("Request: %+v\n", r)
	log.Printf("Body: %+v\n", u)

	w.WriteHeader(404)
	fmt.Fprintf(w, "Page not found - 404\n")
}

func main() {
	var db backend.Backend

	// Get user options
	// 	port    - default to 8443
	// 	backend - default to random
	portPtr := flag.Int("port", 8443, "server port")
	backendPtr := flag.String("backend", "random", "the backend to use [random, sqlite]")
	flag.Parse()

	// Hreate and init the backend
	if *backendPtr == "sqlite" {
		db = &backend.Sqlite{}
	} else {
		db = &backend.Random{}
	}
	db.Open()

	// h common variables to be used by all Handler functions
	// backend the backend to use for metrics source
	// version the Hawkular server version we mimic
	h := Handler{
		backend: db,
		version: "0.21.0",
	}

	// Create the routers
	// Requests no handled by the routers will be forworded to BadRequest Handler
	rRoot := router.Router{
		Prefix: "/",
		Next:   BadRequest{},
	}
	// Root Routing table
	rRoot.Add("GET", "oapi", h.GetAPIVersions)

	rAlerts := router.Router{
		Prefix: "/hawkular/alerts/",
		Next:   rRoot,
	}
	// Alerts Routing table
	rAlerts.Add("GET", "status", h.GetStatus)

	rMetrics := router.Router{
		Prefix: "/hawkular/metrics/",
		Next:   rAlerts,
	}
	// Metrics Routing table
	rMetrics.Add("GET", "status", h.GetStatus)
	rMetrics.Add("GET", "metrics", h.GetMetrics)
	rMetrics.Add("GET", "gauges/:id/raw", h.GetData)
	rMetrics.Add("GET", "counters/:id/raw", h.GetData)
	rMetrics.Add("GET", "availability/:id/raw", h.GetData)
	rMetrics.Add("GET", "gauges/:id/stats", h.GetData)
	rMetrics.Add("POST", "gauges/raw", h.PostData)
	rMetrics.Add("PUT", "gauges/:id/tags", h.PutTags)

	// Run the server
	srv := &http.Server{
		Addr:           fmt.Sprintf("0.0.0.0:%d", *portPtr),
		Handler:        rMetrics,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Printf("Start server, listen on https://%+v", srv.Addr)
	log.Fatal(srv.ListenAndServeTLS("server.pem", "server.key"))
}
