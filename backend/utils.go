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
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// validRegex regexp for validating sql variables
var validRegex = regexp.MustCompile(`^[ A-Za-z0-9_,|:/\[\]\(\)\.\+\*-]*$`)

// json struct used to query data by the POST http request
type dataQuery struct {
	IDs            []string    `json:"ids"`
	Start          json.Number `json:"start"`
	End            json.Number `json:"end"`
	Limit          json.Number `json:"limit"`
	Order          string      `json:"order"`
	BucketDuration string      `json:"bucketDuration"`
}

// json struct used to parse post data http request
type postDataItem struct {
	Timestamp json.Number `json:"timestamp"`
	Value     json.Number `json:"value"`
}
type postDataItems struct {
	ID   string         `json:"id"`
	Data []postDataItem `json:"data"`
}

// json struct used to parse put tags http request
type putTags struct {
	ID   string            `json:"id"`
	Tags map[string]string `json:"tags"`
}

// getData querys data from the backend, and return a json string
func getData(h Handler, tenant string, id string, end int64, start int64, limit int64, order string, bucketDuration int64) string {
	var resStr string

	// call backend for data
	if bucketDuration == 0 {
		res := h.Backend.GetRawData(tenant, id, end, start, limit, order)
		resJSON, _ := json.Marshal(res)
		resStr = string(resJSON)
	} else {
		res := h.Backend.GetStatData(tenant, id, end, start, limit, order, bucketDuration)
		resJSON, _ := json.Marshal(res)
		resStr = string(resJSON)
	}

	return resStr
}

// parseTags takes a comma separeted key:value list string and returns a map[string]string
// 	e.g.
// 	"warm:kitty,soft:kitty" => {"warm": "kitty", "soft": "kitty"}
func parseTags(tags string) map[string]string {
	vsf := make(map[string]string)

	tagsList := strings.Split(tags, ",")
	for _, tag := range tagsList {
		t := strings.Split(tag, ":")
		if len(t) == 2 {
			vsf[t[0]] = t[1]
		}
	}
	return vsf
}

func validStr(s string) bool {
	valid := validRegex.MatchString(s)
	if !valid {
		log.Printf("Valid string fail: %s\n", s)
	}
	return valid
}

func validTags(tags map[string]string) bool {
	for k, v := range tags {
		if !validStr(k) || !validStr(v) {
			return false
		}
	}

	return true
}

func FilterItems(vs []Item, f func(Item) bool) []Item {
	vsf := make([]Item, 0)

	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

// parseTenant return the tenant header value or "_ops"
func parseTenant(r *http.Request) string {
	tenant := r.Header.Get("Hawkular-Tenant")
	if tenant == "" {
		tenant = "_ops"
	}

	return tenant
}

func parseTimespan(r *http.Request) (int64, int64, int64) {
	end := int64(time.Now().UTC().Unix() * 1000)
	if v, ok := r.Form["end"]; ok && len(v) > 0 {
		if i, err := strconv.Atoi(v[0]); err == nil && i > 1 {
			end = int64(i)
		}
	}

	start := end - int64(8*60*60*1000)
	if v, ok := r.Form["start"]; ok && len(v) > 0 {
		if i, err := strconv.Atoi(v[0]); err == nil && i > 1 {
			start = int64(i)
		}
	}

	bucketDuration := int64(0)
	if v, ok := r.Form["bucketDuration"]; ok && len(v) > 0 {
		if i, err := strconv.Atoi(v[0][:len(v[0])-1]); err == nil && i > 1 {
			bucketDuration = int64(i)
		}
	}

	return end, start, bucketDuration
}
