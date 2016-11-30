package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/yaacov/mohawk/backends"
	"github.com/yaacov/mohawk/router"
)

const VER = "0.1.0"

func main() {
	var db backend.Backend

	portPtr := flag.Int("port", 8443, "server port")
	backendPtr := flag.String("backend", "random", "the backend to use [random, sqlite]")
	flag.Parse()

	if *backendPtr == "sqlite" {
		db = &backend.Sqlite{}
	} else {
		db = &backend.Random{}
	}
	db.Open()

	h := Handler{
		backend: db,
		version: "0.21.0",
	}
	r := router.Router{
		Prefix:           "/",
		HandleBadRequest: h.BadRequest,
	}

	r.Add("GET", "oapi", h.GetAPIVersions)
	r.Add("GET", "hawkular/metrics/status", h.GetStatus)

	r.Add("GET", "hawkular/metrics/metrics", h.GetMetrics)
	r.Add("GET", "hawkular/metrics/gauges/:id/raw", h.GetData)
	r.Add("GET", "hawkular/metrics/counters/:id/raw", h.GetData)
	r.Add("GET", "hawkular/metrics/availability/:id/raw", h.GetData)

	r.Add("GET", "hawkular/metrics/gauges/:id/stats", h.GetData)
	r.Add("GET", "hawkular/metrics/counters/:id/stats", h.GetData)

	r.Add("POST", "hawkular/metrics/gauges/raw", h.PostData)
	r.Add("PUT", "hawkular/metrics/gauges/:id/tags", h.PutTags)
	r.Add("PUT", "hawkular/metrics/counters/:id/tags", h.PutTags)

	srv := &http.Server{
		Addr:           fmt.Sprintf("0.0.0.0:%d", *portPtr),
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Printf("Start server, listen on https://%+v", srv.Addr)
	log.Fatal(srv.ListenAndServeTLS("server.pem", "server.key"))
}
