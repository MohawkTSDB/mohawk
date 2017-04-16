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
	"net/url"
	"time"

	"github.com/yaacov/mohawk/backend"
	"github.com/yaacov/mohawk/backend/random"
	"github.com/yaacov/mohawk/backend/sqlite"
	"github.com/yaacov/mohawk/backend/timeout"
	"github.com/yaacov/mohawk/middleware"
	"github.com/yaacov/mohawk/router"
)

// VER the server version
const VER = "0.8.3"

// defaults
const defaultPort = 8080
const defaultBackend = "sqlite"
const defaultAPI = "0.21.0"
const defaultTLS = "false"

// ImplementationVersion Hakular server api implementation version
var ImplementationVersion string

// BackendName MoHawk active backend
var BackendName string

func main() {
	var db backend.Backend

	// Get user options
	portPtr := flag.Int("port", defaultPort, "server port")
	backendPtr := flag.String("backend", defaultBackend, "the backend to use [random, sqlite, timeout]")
	apiPtr := flag.String("api", defaultAPI, "the hawkulr api to mimic [e.g. 0.8.9.Testing, 0.21.2.Final]")
	tlsPtr := flag.String("tls", defaultTLS, "use TLS server")
	optionsPtr := flag.String("options", "", "specific backend options")
	verbosePtr := flag.Bool("verbose", false, "more debug output")
	versionPtr := flag.Bool("version", false, "version number")
	flag.Parse()

	// return version number and exit
	if *versionPtr {
		fmt.Printf("MoHawk version: %s\n\n", VER)
		return
	}

	// Create and init the backend
	switch *backendPtr {
	case "sqlite":
		db = &sqlite.Backend{}
	case "timeout":
		db = &timeout.Backend{}
	case "random":
		db = &random.Backend{}
	default:
		log.Fatal("Can't find backend:", *backendPtr)
	}

	// parse options
	if options, err := url.ParseQuery(*optionsPtr); err == nil {
		db.Open(options)
	} else {
		log.Fatal("Can't parse opetions:", *optionsPtr)
	}

	// set global variables
	ImplementationVersion = *apiPtr
	BackendName = db.Name()

	// h common variables to be used for the backend Handler functions
	// backend the backend to use for metrics source
	h := backend.Handler{
		Verbose: *verbosePtr,
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
	rRoot.Add("GET", "hawkular/metrics/tenants", GetTenants)

	rMetrics := router.Router{
		Prefix: "/hawkular/metrics/",
	}
	// Metrics Routing table
	if *backendPtr == "timeout" {
		rMetrics.Add("GET", "metrics", GetTimeout)
	} else {
		rMetrics.Add("GET", "metrics", h.GetMetrics)

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
	}

	// logger a logging middleware
	logger := middleware.Logger{
		Verbose: *verbosePtr,
	}

	// gzipper a gzip encoding middleware
	gzipper := middleware.GZipper{
		Verbose: *verbosePtr,
	}

	// fallback a BadRequest middleware
	fallback := middleware.BadRequest{}

	// concat middlewars and routes (first logger until rRoot) with a fallback to BadRequest
	middlewareList := []middleware.MiddleWare{&logger, &gzipper, &rMetrics, &rRoot, &fallback}
	middleware.Append(middlewareList)

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
