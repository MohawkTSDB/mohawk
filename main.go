// Copyright 2016 Red Hat, Inc. and/or its affiliates
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

// Package main
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/yaacov/mohawk/backend"
	"github.com/yaacov/mohawk/middleware"
	"github.com/yaacov/mohawk/router"
)

// VER the server version
const VER = "0.8.3"

// ImplementationVersion Hakular server api implementation version
var ImplementationVersion string

// BackendName MoHawk active backend
var BackendName string

func main() {
	var db backend.Backend

	// Get user options
	// 	port    - default to 8443
	// 	backend - default to random
	portPtr := flag.Int("port", 8443, "server port")
	backendPtr := flag.String("backend", "random", "the backend to use [random, sqlite, error]")
	apiPtr := flag.String("api", "0.21.0", "the hawkulr api to mimic [e.g. 0.8.9.Testing, 0.21.2.Final]")
	tlsPtr := flag.String("tls", "true", "use TLS server")
	versionPtr := flag.Bool("version", false, "version number")
	flag.Parse()

	// return version number and exit
	if *versionPtr {
		fmt.Printf("MoHawk version: %s\n\n", VER)
		return
	}

	// Create and init the backend
	if *backendPtr == "sqlite" {
		db = &backend.Sqlite{}
	} else if *backendPtr == "error" {
		db = &backend.Timeout{}
	} else if *backendPtr == "random" {
		db = &backend.Random{}
	} else {
		log.Fatal("Can't find backend:", *backendPtr)
	}
	db.Open()

	// set global variables
	ImplementationVersion = *apiPtr
	BackendName = db.Name()

	// h common variables to be used by all Handler functions
	// backend the backend to use for metrics source
	h := backend.Handler{
		Backend: db,
	}

	// Create the routers
	// Requests not handled by the routers will be forworded to BadRequest Handler
	rRoot := router.Router{
		Prefix: "/",
	}
	// Root Routing table
	rRoot.Add("GET", "oapi", GetAPIVersions)
	rRoot.Add("GET", "hawkular/metrics/status", GetStatus)

	rTimeout := router.Router{
		Prefix: "/hawkular/metrics/",
	}
	// Timeout Error Routing table
	rTimeout.Add("GET", "metrics", Timeout)
	rTimeout.Add("GET", "tenants", h.GetTenants)

	rMetrics := router.Router{
		Prefix: "/hawkular/metrics/",
	}
	// Metrics Routing table
	rMetrics.Add("GET", "metrics", h.GetMetrics)
	rMetrics.Add("GET", "tenants", h.GetTenants)

	// api version >= 0.16.0
	rMetrics.Add("GET", "gauges/:id/raw", h.GetData)
	rMetrics.Add("GET", "counters/:id/raw", h.GetData)
	rMetrics.Add("GET", "availability/:id/raw", h.GetData)

	rMetrics.Add("GET", "gauges/:id/stats", h.GetData)
	rMetrics.Add("GET", "counters/:id/stats", h.GetData)
	rMetrics.Add("GET", "availability/:id/stats", h.GetData)

	rMetrics.Add("POST", "gauges/raw", h.PostData)
	rMetrics.Add("POST", "gauges/raw/query", h.PostQuery)
	rMetrics.Add("POST", "counters/raw", h.PostData)
	rMetrics.Add("POST", "counters/raw/query", h.PostQuery)

	rMetrics.Add("PUT", "gauges/:id/tags", h.PutTags)
	rMetrics.Add("PUT", "counters/:id/tags", h.PutTags)

	// api version < 0.16.0
	rMetrics.Add("GET", "gauges/:id/data", h.GetData)
	rMetrics.Add("GET", "counters/:id/data", h.GetData)
	rMetrics.Add("GET", "availability/:id/data", h.GetData)

	rMetrics.Add("POST", "gauges/data", h.PostData)
	rMetrics.Add("POST", "counters/data", h.PostData)

	// logger a logging middleware
	logger := middleware.Logger{}

	// gzipper a gzip encoding middleware
	gzipper := middleware.GZipper{}

	// fallback a BadRequest middleware
	fallback := middleware.BadRequest{}

	// concat middlewars and routes (first logger until rRoot) with a fallback to BadRequest
	if *backendPtr == "error" {
		middlewareList := []middleware.MiddleWare{&logger, &gzipper, &rTimeout, &rRoot, &fallback}
		middleware.Append(middlewareList)
	} else {
		middlewareList := []middleware.MiddleWare{&logger, &gzipper, &rMetrics, &rRoot, &fallback}
		middleware.Append(middlewareList)
	}

	// Run the server
	srv := &http.Server{
		Addr:           fmt.Sprintf("0.0.0.0:%d", *portPtr),
		Handler:        logger,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if *tlsPtr == "true" {
		log.Printf("Start server, listen on https://%+v", srv.Addr)
		log.Fatal(srv.ListenAndServeTLS("server.pem", "server.key"))
	} else {
		log.Printf("Start server, listen on http://%+v", srv.Addr)
		log.Fatal(srv.ListenAndServe())
	}
}
