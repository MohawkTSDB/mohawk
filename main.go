package main

import (
	"log"
	"net/http"
	"time"

	"github.com/yaacov/mohawk/backends"
	"github.com/yaacov/mohawk/router"
)

func main() {
	b := backend.Random{}
	b.Open()
	h := Handler{backend: b}
	r := router.Router{
		Prefix:           "/hawkular/metrics/",
		HandleBadRequest: h.handleBadRequest}

	r.Add("GET", "status", h.handleStatus)
	r.Add("GET", "metrics", h.handleMetrics)
	r.Add("GET", "gauges/:id/raw", h.handleGetData)

	srv := &http.Server{
		Addr:           ":8443",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(srv.ListenAndServeTLS("server.pem", "server.key"))
}
