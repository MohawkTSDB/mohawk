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

// Package server API REST server
package server

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/spf13/viper"

	"github.com/MohawkTSDB/mohawk/src/alerts"
	"github.com/MohawkTSDB/mohawk/src/middleware"
	"github.com/MohawkTSDB/mohawk/src/router"
	"github.com/MohawkTSDB/mohawk/src/server/handlers"
	"github.com/MohawkTSDB/mohawk/src/storage"
	"github.com/MohawkTSDB/mohawk/src/storage/example"
	"github.com/MohawkTSDB/mohawk/src/storage/memory"
	"github.com/MohawkTSDB/mohawk/src/storage/mongo"
	"github.com/MohawkTSDB/mohawk/src/storage/sqlite"
)

// VER the server version
const VER = "0.27.1"

// defaults
const defaultAPI = "0.21.0"
const publicPath = "^/hawkular/metrics/status$"

// BackendName Mohawk active storage
var BackendName string

// GetStatus return a json status struct
func GetStatus(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	resTemplate := `{"MetricsService":"STARTED","Implementation-Version":"%s","MohawkVersion":"%s","MohawkBackend":"%s"}`
	res := fmt.Sprintf(resTemplate, defaultAPI, VER, BackendName)

	w.WriteHeader(200)
	fmt.Fprintln(w, res)
}

// Serve run the REST API server
func Serve() error {
	var db storage.Storage
	var alertRules *alerts.AlertRules
	var routers http.HandlerFunc
	var authorizationKey string

	var backendQuery = viper.GetString("storage")
	var optionsQuery = viper.GetString("options")
	var verbose = viper.GetBool("verbose")
	var media = viper.GetString("media")
	var tls = viper.GetBool("tls")
	var gzip = viper.GetBool("gzip")
	var bearerAuth = viper.GetString("bearer-auth")
	var basicAuth = viper.GetString("basic-auth")
	var port = viper.GetInt("port")
	var cert = viper.GetString("cert")
	var key = viper.GetString("key")
	var alertsInterval = viper.GetInt("alerts-interval")
	var alertsServerURL = viper.GetString("alerts-server")
	var alertsServerMethod = viper.GetString("alerts-server-method")
	var alertsServerInsecure = viper.GetBool("alerts-server-insecure")
	var configAlerts = viper.ConfigFileUsed() != "" && viper.Get("alerts") != ""

	// Create and init the storage
	switch backendQuery {
	case "sqlite":
		db = &sqlite.Storage{}
	case "memory":
		db = &memory.Storage{}
	case "mongo":
		db = &mongo.Storage{}
	case "example":
		db = &example.Storage{}
	default:
		log.Fatal("Can't find storage:", backendQuery)
	}

	// parse options
	if options, err := url.ParseQuery(optionsQuery); err == nil {
		db.Open(options)
	} else {
		log.Fatal("Can't parse opetions:", optionsQuery)
	}

	// set global variables
	BackendName = db.Name()

	// Create alerts runner
	if configAlerts {
		// parse alert list from config yaml
		l := []*alerts.Alert{}
		viper.UnmarshalKey("alerts", &l)

		if len(l) > 0 {
			// creat and Init the alert handler
			alertRules = &alerts.AlertRules{
				Storage:        db,
				Verbose:        verbose,
				Alerts:         l,
				AlertsInterval: alertsInterval,
				ServerURL:      alertsServerURL,
				ServerMethod:   alertsServerMethod,
				ServerInsecure: alertsServerInsecure,
			}
			alertRules.Init()
		}
	}

	// h common variables to be used for the storage Handler functions
	// Storage the storage to use for metrics source
	h := APIHhandler{
		Verbose: verbose,
		Storage: db,
		Alerts:  alertRules,
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
	rGauges.Add("POST", "stats/query", h.PostQuery)

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
	rCounters.Add("POST", "stats/query", h.PostQuery)

	rAvailability := router.Router{
		Prefix: "/hawkular/metrics/availability/",
	}
	rAvailability.Add("GET", ":id/raw", h.GetData)
	rAvailability.Add("GET", ":id/stats", h.GetData)

	// Requests not handled by the routers will be forworded to BadRequest Handler
	rAlerts := router.Router{
		Prefix: "/hawkular/alerts/",
	}
	rAlerts.Add("GET", "status", h.GetAlertsStatus)
	rAlerts.Add("GET", "raw", h.GetAlerts)

	// Create the http handlers
	// logging handler
	logger := handler.Logger{}

	// set the authorization key
	if basicAuth != "" {
		authorizationKey = "Basic " + base64.StdEncoding.EncodeToString([]byte(basicAuth))
	}

	if bearerAuth != "" {
		authorizationKey = "Bearer " + bearerAuth
	}

	// add headers to response
	headers := handler.Headers{}

	// static a file server handler
	static := handler.Static{
		MediaPath: media,
	}

	// badrequest a BadRequest handler
	badrequest := handler.BadRequest{
		Verbose: verbose,
	}

	// concat all routers and add fallback handler
	if authorizationKey == "" {
		routers = handler.Append(
			&logger, &headers, &rGauges, &rCounters, &rAvailability, &rAlerts, &rRoot, &static, &badrequest)
	} else {
		// create an authentication handler
		authorization := handler.Authorization{
			// init the values:
			//    make the public path a regex once on init
			//    create the full Authorization header using the bearer token
			// this will prevent this values from re-calculate each http request
			PublicPathRegex: regexp.MustCompile(publicPath),
			Authorization:   authorizationKey,
			Verbose:         verbose,
		}

		routers = handler.Append(
			&logger, &authorization, &headers, &rGauges, &rCounters, &rAvailability, &rAlerts, &rRoot, &static, &badrequest)
	}

	// Create a list of middlwares
	decorators := []middleware.Decorator{}
	if gzip {
		decorators = append(decorators, middleware.GzipDecodeDecorator(), middleware.GzipEncodeDecorator())
	}

	// concat middlewars and routes (first logger until rRoot) with a fallback to BadRequest
	core := middleware.Append(routers, decorators...)

	// Run the server
	srv := &http.Server{
		Addr:           fmt.Sprintf("0.0.0.0:%d", port),
		Handler:        core,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if tls {
		log.Printf("Start server, listen on https://%+v", srv.Addr)
		log.Fatal(srv.ListenAndServeTLS(cert, key))
	} else {
		log.Printf("Start server, listen on http://%+v", srv.Addr)
		log.Fatal(srv.ListenAndServe())
	}

	return nil
}
