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
	"net/http"
	"strconv"
	"time"
)

// Handler common variables to be used by all Handler functions
// 	version the version of the Hawkular server we are mocking
// 	backend the backend to be used by the Handler functions
type Handler struct {
	Backend Backend
}

// GetMetrics return a list of metrics definitions
func (h Handler) GetMetrics(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	var res []Item

	r.ParseForm()

	// we only use gauges
	if typeStr, ok := r.Form["type"]; ok && len(typeStr) > 0 && typeStr[0] != "gauge" {
		w.WriteHeader(200)
		fmt.Fprintln(w, "[]")

		return
	}

	// get a list of gauges
	if tagsStr, ok := r.Form["tags"]; ok && len(tagsStr) > 0 {
		tags := parseTags(tagsStr[0])
		if !validTags(tags) {
			w.WriteHeader(504)
			return
		}
		res = h.Backend.GetItemList(tags)
	} else {
		res = h.Backend.GetItemList(map[string]string{})
	}
	resJSON, _ := json.Marshal(res)

	w.WriteHeader(200)
	fmt.Fprintln(w, string(resJSON))
}

// GetData return a list of metrics raw / stat data
func (h Handler) GetData(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	var resStr string
	var id string
	var end int64
	var start int64
	var limit int64
	var bucketDuration int64
	var order string

	// use the id from the argv list
	id = argv["id"]
	if !validStr(id) {
		w.WriteHeader(504)
		return
	}

	// get data from the form arguments
	r.ParseForm()
	if v, ok := r.Form["end"]; ok && len(v) > 0 {
		i, _ := strconv.Atoi(v[0])
		end = int64(i)
	} else {
		end = int64(time.Now().Unix() * 1000)
	}
	if v, ok := r.Form["start"]; ok && len(v) > 0 {
		i, _ := strconv.Atoi(v[0])
		start = int64(i)
	} else {
		start = end - int64(8*60*60*1000)
	}
	if v, ok := r.Form["limit"]; ok && len(v) > 0 {
		i, _ := strconv.Atoi(v[0])
		limit = int64(i)
		if limit < 1 {
			limit = int64(20000)
		}
	} else {
		limit = int64(20000)
	}
	if v, ok := r.Form["order"]; ok && len(v) > 0 {
		order = v[0]
		// do sanity check
		if order != "ASC" || order != "DESC" {
			order = "ASC"
		}
	} else {
		order = "ASC"
	}
	if v, ok := r.Form["bucketDuration"]; ok && len(v) > 0 {
		i, _ := strconv.Atoi(v[0][:len(v[0])-1])
		bucketDuration = int64(i)
	} else {
		bucketDuration = int64(0)
	}

	// call backend for data
	if bucketDuration == 0 {
		res := h.Backend.GetRawData(id, end, start, limit, order)
		resJSON, _ := json.Marshal(res)
		resStr = string(resJSON)
	} else {
		res := h.Backend.GetStatData(id, end, start, limit, order, bucketDuration)
		resJSON, _ := json.Marshal(res)
		resStr = string(resJSON)
	}

	// output to client
	w.WriteHeader(200)
	fmt.Fprintln(w, resStr)
}

// PostQuery send timestamp, value to the backend
func (h Handler) PostQuery(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	var resStr string
	var u dataQuery

	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()
	decoder.Decode(&u)

	id := u.IDs[0]

	if !validStr(id) {
		w.WriteHeader(504)
		return
	}

	end, _ := u.End.Int64()
	if end < 1 {
		end = int64(time.Now().Unix() * 1000)
	}

	start, _ := u.Start.Int64()
	if start < 1 {
		start = end - int64(8*60*60*1000)
	}

	limit, _ := u.Start.Int64()
	if limit < 1 {
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

	// call backend for data
	if bucketDuration == 0 {
		res := h.Backend.GetRawData(id, end, start, limit, order)
		resJSON, _ := json.Marshal(res)
		resStr = string(resJSON)
	} else {
		res := h.Backend.GetStatData(id, end, start, limit, order, bucketDuration)
		resJSON, _ := json.Marshal(res)
		resStr = string(resJSON)
	}

	// output to client
	w.WriteHeader(200)
	fmt.Fprintf(w, "[{\"id\": \"%s\", \"data\": %s}]", id, resStr)
}

// PostData send timestamp, value to the backend
func (h Handler) PostData(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	var u []map[string]interface{}
	json.NewDecoder(r.Body).Decode(&u)

	id := u[0]["id"].(string)
	if !validStr(id) {
		w.WriteHeader(504)
		return
	}

	t := u[0]["data"].([]interface{})[0].(map[string]interface{})["timestamp"].(float64)
	vStr := u[0]["data"].([]interface{})[0].(map[string]interface{})["value"].(string)
	v, _ := strconv.ParseFloat(vStr, 64)

	h.Backend.PostRawData(id, int64(t), v)
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

	h.Backend.PutTags(id, tags)
	w.WriteHeader(200)
	fmt.Fprintln(w, "{}")
}
