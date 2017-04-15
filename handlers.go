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
	"fmt"
	"net/http"
)

// GetStatus return a json status struct
func GetStatus(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	resTemplate := `{
	"MetricsService":"STARTED",
	"Implementation-Version":"%s",
	"MohawkVersion":"%s",
	"MohawkBackend":"%s"
}`
	res := fmt.Sprintf(resTemplate, ImplementationVersion, VER, BackendName)

	w.WriteHeader(200)
	fmt.Fprintln(w, res)
}

// GetAPIVersions return a json apiVersion struct
func GetAPIVersions(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	resTemplate := `{
	"kind": "APIVersions",
	"apiVersion": "%s",
	"versions": [
		"%s"
	],
	"serverAddressByClientCIDRs": null
}`
	res := fmt.Sprintf(resTemplate, "v1", "v1")

	w.WriteHeader(200)
	fmt.Fprintln(w, res)
}

// Timeout a timeout 504 Error
func Timeout(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	res := "<html><body><h1>504 Gateway Time-out</h1>The server didn't respond in time.</body></html>"

	w.WriteHeader(504)
	fmt.Fprintln(w, res)
}
