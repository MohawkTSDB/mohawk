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

// Package backend define the Backend interface
package backend

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Handler common variables to be used by all Handler functions
// 	version the version of the Hawkular server we are mocking
// 	backend the backend to be used by the Handler functions
type Handler struct {
	Verbose bool
	Backend Backend
}

// GetTenants return a list of metrics tenants
func (h Handler) GetTenants(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	var res []Tenant

	res = h.Backend.GetTenants()
	resJSON, _ := json.Marshal(res)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintln(w, string(resJSON))
}

// GetMetrics return a list of metrics definitions
func (h Handler) GetMetrics(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	var res []Item

	r.ParseForm()

	if h.Verbose {
		log.Printf("Metrics type: %s", r.Form.Get("type"))
	}

	// we only use gauges
	if typeStr, ok := r.Form["type"]; ok && len(typeStr) > 0 && typeStr[0] != "gauge" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		fmt.Fprintln(w, "[]")

		return
	}

	// get tenant
	tenant := parseTenant(r)

	// get a list of gauges
	if tagsStr, ok := r.Form["tags"]; ok && len(tagsStr) > 0 {
		tags := parseTags(tagsStr[0])
		if !validTags(tags) {
			w.WriteHeader(504)
			return
		}
		res = h.Backend.GetItemList(tenant, tags)
	} else {
		res = h.Backend.GetItemList(tenant, map[string]string{})
	}
	resJSON, _ := json.Marshal(res)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintln(w, string(resJSON))
}

// GetData return a list of metrics raw / stat data
func (h Handler) GetData(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	// use the id from the argv list
	id := argv["id"]
	if !validStr(id) {
		w.WriteHeader(504)
		return
	}

	// get data from the form arguments
	r.ParseForm()

	// get tenant
	tenant := parseTenant(r)

	// get timespan
	end, start, bucketDuration := parseTimespan(r)

	limit := int64(20000)
	if v, ok := r.Form["limit"]; ok && len(v) > 0 {
		if i, err := strconv.Atoi(v[0]); err == nil && i > 0 {
			limit = int64(i)
		}
	}

	order := "DESC"
	if v, ok := r.Form["order"]; ok && len(v) > 0 && v[0] == "ASC" {
		order = "ASC"
	}

	if h.Verbose {
		log.Printf("ID: %s@%s, End: %d, Start: %d, Limit: %d, Order: %s, bucketDuration: %ds", tenant, id, end, start, limit, order, bucketDuration)
	}

	// call backend for data
	resStr := getData(h, tenant, id, end, start, limit, order, bucketDuration)

	// output to client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintf(w, resStr)
}

// DeleteData delete a list of metrics raw  data
func (h Handler) DeleteData(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	// use the id from the argv list
	id := argv["id"]
	if !validStr(id) {
		w.WriteHeader(504)
		return
	}

	// get data from the form arguments
	r.ParseForm()

	// get tenant
	tenant := parseTenant(r)

	// get timespan
	end, start, _ := parseTimespan(r)

	if h.Verbose {
		log.Printf("ID: %s@%s, End: %d, Start: %d", tenant, id, end, start)
	}

	// call backend for data
	if start < end {
		h.Backend.DeleteData(tenant, id, end, start)

		// output to client
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		fmt.Fprintf(w, "{}")
		return
	}

	w.WriteHeader(504)
	fmt.Fprintf(w, "504 - Can't delete time rage")
}

// PostQuery send timestamp, value to the backend
func (h Handler) PostQuery(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	var u dataQuery
	var end int64
	var start int64
	var limit int64
	var err error

	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()
	decoder.Decode(&u)

	// get tenant
	tenant := parseTenant(r)

	for _, id := range u.IDs {
		if !validStr(id) {
			w.WriteHeader(504)
			fmt.Fprintf(w, "<p>Error 504, Bad metrics ID</p>")
			return
		}
	}
	numOfItems := len(u.IDs) - 1

	if end, err = u.End.Int64(); err != nil || end < 1 {
		end = int64(time.Now().UTC().Unix() * 1000)
	}

	if start, err = u.Start.Int64(); err != nil || start < 1 {
		start = end - int64(8*60*60*1000)
	}

	if limit, err = u.Limit.Int64(); err != nil || limit < 1 {
		limit = int64(20000)
	}

	order := "ASC"
	if u.Order == "DESC" {
		order = "DESC"
	}

	bucketDuration := int64(0)
	if v := u.BucketDuration; len(v) > 1 {
		if i, err := strconv.Atoi(v[:len(v)-1]); err == nil {
			bucketDuration = int64(i)
		}
	}

	if h.Verbose {
		log.Printf("Tenant: %s, IDs: %+v", tenant, u.IDs)
		log.Printf("End: %d, Start: %d, Limit: %d, Order: %s, bucketDuration: %ds", end, start, limit, order, bucketDuration)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintf(w, "[")

	for i, id := range u.IDs {
		// call backend for data
		resStr := getData(h, tenant, id, end, start, limit, order, bucketDuration)

		// write data
		fmt.Fprintf(w, "{\"id\": \"%s\", \"data\": %s}", id, resStr)
		if i < numOfItems {
			fmt.Fprintf(w, ",")
		}
	}

	fmt.Fprintf(w, "]")
}

// PostData send timestamp, value to the backend
func (h Handler) PostData(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	var u []postDataItems
	json.NewDecoder(r.Body).Decode(&u)

	for _, item := range u {
		if !validStr(item.ID) {
			w.WriteHeader(504)
			fmt.Fprintf(w, "<p>Error 504, Bad metrics ID</p>")
			return
		}
	}

	// get tenant
	tenant := parseTenant(r)

	for _, item := range u {
		id := item.ID

		for _, data := range item.Data {
			timestamp, _ := data.Timestamp.Int64()
			value, _ := data.Value.Float64()

			h.Backend.PostRawData(tenant, id, timestamp, value)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintln(w, "{}")
}

// PutTags send tag, value pairs to the backend
func (h Handler) PutTags(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	var tags map[string]string
	json.NewDecoder(r.Body).Decode(&tags)

	// use the id from the argv list
	id := argv["id"]
	if !validStr(id) || !validTags(tags) {
		w.WriteHeader(504)
		return
	}

	// get tenant
	tenant := parseTenant(r)

	h.Backend.PutTags(tenant, id, tags)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintln(w, "{}")
}

// PutMultiTags send tags pet dataItem - tag, value pairs to the backend
func (h Handler) PutMultiTags(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	var u []putTags
	json.NewDecoder(r.Body).Decode(&u)

	for _, item := range u {
		if !validStr(item.ID) {
			w.WriteHeader(504)
			fmt.Fprintf(w, "<p>Error 504, Bad metrics ID</p>")
			return
		}
	}

	// get tenant
	tenant := parseTenant(r)

	for _, item := range u {
		id := item.ID
		if validTags(item.Tags) {
			h.Backend.PutTags(tenant, id, item.Tags)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintln(w, "{}")
}

// DeleteTags delete a tag
func (h Handler) DeleteTags(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	// use the id from the argv list
	id := argv["id"]
	tagsStr := argv["tags"]
	if !validStr(id) || !validStr(tagsStr) {
		w.WriteHeader(504)
		return
	}
	tags := strings.Split(tagsStr, ",")

	// get tenant
	tenant := parseTenant(r)

	h.Backend.DeleteTags(tenant, id, tags)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintln(w, "{}")
}
