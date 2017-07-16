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

// Package main
package main

import (
	"fmt"
	"github.com/urfave/cli"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/yaacov/mohawk/backend"
	"github.com/yaacov/mohawk/backend/example"
	"github.com/yaacov/mohawk/backend/memory"
	"github.com/yaacov/mohawk/backend/mongo"
	"github.com/yaacov/mohawk/backend/sqlite"
	"github.com/yaacov/mohawk/middleware"
	"github.com/yaacov/mohawk/router"
)

// VER the server version
const VER = "0.19.5"

// defaults
const defaultPort = 8080
const defaultBackend = "sqlite"
const defaultAPI = "0.21.0"
const defaultTLS = false
const defaultTLSKey = "server.key"
const defaultTLSCert = "server.pem"

// BackendName Mohawk active backend
var BackendName string

// GetStatus return a json status struct
func GetStatus(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	resTemplate := `{"MetricsService":"STARTED","Implementation-Version":"%s","MohawkVersion":"%s","MohawkBackend":"%s"}`
	res := fmt.Sprintf(resTemplate, defaultAPI, VER, BackendName)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintln(w, res)
}

func main() {
	app := cli.NewApp()
	app.Name = "mohawk"
	app.Version = VER
	app.Usage = "Metric data storage engine"
	app.Authors = []cli.Author{
		{Name: "Yaacov Zamir", Email: "kobi.zamir@gmail.com"},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "backend,b", Value: "memory", Usage: "the backend plugin to use"},
		cli.StringFlag{Name: "token", Value: "", Usage: "authorization token"},
		cli.StringFlag{Name: "key", Value: defaultTLSKey, Usage: "path to TLS key file"},
		cli.StringFlag{Name: "cert", Value: defaultTLSCert, Usage: "path to TLS cert file"},
		cli.StringFlag{Name: "options", Value: "", Usage: "specific backend options [e.g. db-dirname, db-url]"},
		cli.UintFlag{Name: "port,p", Value: defaultPort, Usage: "server port"},
		cli.BoolFlag{Name: "tls,t", Usage: "use TLS server"},
		cli.BoolFlag{Name: "gzip,g", Usage: "enable gzip encoding"},
		cli.BoolFlag{Name: "verbose,V", Usage: "more debug output"},
	}
	app.Action = serve
	app.Run(os.Args)
}

func serve(c *cli.Context) error {
	var db backend.Backend
	var middlewareList []middleware.MiddleWare

	// Create and init the backend
	switch c.String("backend") {
	case "sqlite":
		db = &sqlite.Backend{}
	case "memory":
		db = &memory.Backend{}
	case "mongo":
		db = &mongo.Backend{}
	case "example":
		db = &example.Backend{}
	default:
		log.Fatal("Can't find backend:", c.String("backend"))
	}

	// parse options
	if options, err := url.ParseQuery(c.String("options")); err == nil {
		db.Open(options)
	} else {
		log.Fatal("Can't parse opetions:", c.String("options"))
	}

	// set global variables
	BackendName = db.Name()

	// h common variables to be used for the backend Handler functions
	// backend the backend to use for metrics source
	h := backend.Handler{
		Verbose: c.Bool("verbose"),
		Backend: db,
	}

	// Create the routers
	// Requests not handled by the routers will be forworded to BadRequest Handler
	rRoot := router.Router{
		Prefix: "/hawkular/metrics/",
	}
	// Root Routing table
	rRoot.Add("GET", "status", GetStatus)
	rRoot.Add("GET", "tenants", h.GetTenants)
	rRoot.Add("GET", "metrics", h.GetMetrics)

	// Metrics Routing tables
	rGauges := router.Router{
		Prefix: "/hawkular/metrics/gauges/",
	}
	rGauges.Add("GET", ":id/raw", h.GetData)
	rGauges.Add("GET", ":id/stats", h.GetData)
	rGauges.Add("POST", "raw", h.PostData)
	rGauges.Add("POST", "raw/query", h.PostQuery)
	rGauges.Add("PUT", "tags", h.PutMultiTags)
	rGauges.Add("PUT", ":id/tags", h.PutTags)
	rGauges.Add("DELETE", ":id/raw", h.DeleteData)
	rGauges.Add("DELETE", ":id/tags/:tags", h.DeleteTags)

	// deprecated
	rGauges.Add("GET", ":id/data", h.GetData)
	rGauges.Add("POST", "data", h.PostData)

	rCounters := router.Router{
		Prefix: "/hawkular/metrics/counters/",
	}
	rCounters.Add("GET", ":id/raw", h.GetData)
	rCounters.Add("GET", ":id/stats", h.GetData)
	rCounters.Add("POST", "raw", h.PostData)
	rCounters.Add("POST", "raw/query", h.PostQuery)
	rCounters.Add("PUT", ":id/tags", h.PutTags)

	// deprecated
	rCounters.Add("GET", ":id/data", h.GetData)
	rCounters.Add("POST", "data", h.PostData)

	rAvailability := router.Router{
		Prefix: "/hawkular/metrics/availability/",
	}
	rAvailability.Add("GET", ":id/raw", h.GetData)
	rAvailability.Add("GET", ":id/stats", h.GetData)

	// Create the middlewares
	// logger a logging middleware
	logger := middleware.Logger{
		Verbose: c.Bool("verbose"),
	}

	// authorization middleware
	authorization := middleware.Authorization{
		Verbose:         c.Bool("verbose"),
		UseToken:        c.String("token") != "",
		PublicPathRegex: "^/hawkular/metrics/status$",
		Token:           c.String("token"),
	}

	// gzipper a gzip encoding middleware
	gzipper := middleware.GZipper{
		UseGzip: c.Bool("gzip"),
		Verbose: c.Bool("verbose"),
	}

	// badrequest a BadRequest middleware
	badrequest := middleware.BadRequest{
		Verbose: c.Bool("verbose"),
	}

	// concat middlewars and routes (first logger until rRoot) with a fallback to BadRequest
	middlewareList = []middleware.MiddleWare{
		&logger,
		&authorization,
		&gzipper,
		&rGauges,
		&rCounters,
		&rAvailability,
		&rRoot,
		&badrequest,
	}
	middleware.Append(middlewareList)

	// Run the server
	srv := &http.Server{
		Addr:           fmt.Sprintf("0.0.0.0:%d", c.Int("port")),
		Handler:        logger,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if c.Bool("tls") {
		log.Printf("Start server, listen on https://%+v", srv.Addr)
		log.Fatal(srv.ListenAndServeTLS(c.String("cert"), c.String("key")))
	} else {
		log.Printf("Start server, listen on http://%+v", srv.Addr)
		log.Fatal(srv.ListenAndServe())
	}

	return nil
}
