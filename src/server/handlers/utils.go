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

// Package handler http server handler functions
package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// validRegex regexp for validating sql variables
var validRegex = regexp.MustCompile(`^[ A-Za-z0-9_@,|:/\[\]\(\)\.\+\*-]*$`)

// const defaultStartTime default REST API query start time value
var defaultStartTime = "-8h"

// json struct used to query data by the POST http request
type dataQuery struct {
	IDs            []string          `json:"ids"`
	Tags           map[string]string `json:"tags"`
	End            interface{}       `json:"end"`
	Start          interface{}       `json:"start"`
	BucketDuration interface{}       `json:"bucketDuration"`
	Limit          json.Number       `json:"limit"`
	Order          string            `json:"order"`
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

// ParseTenant return the tenant header value or "_ops"
func parseTenant(r *http.Request) string {
	tenant := r.Header.Get("Hawkular-Tenant")
	if tenant == "" {
		tenant = "_ops"
	}

	return tenant
}

func parseTimespanStrings(e string, s string, b string) (int64, int64, int64) {
	var start int64
	var end int64
	var bucketDuration int64

	if i, err := parseSec(e); err == nil {
		end = i * 1000
	} else {
		end = int64(time.Now().UTC().Unix()) * 1000
	}

	if i, err := parseSec(s); err == nil {
		start = i * 1000
	} else {
		if i, err := parseSec(defaultStartTime); err == nil {
			start = i * 1000
		}
	}

	if i, err := parseSec(b); err == nil {
		bucketDuration = i * 1000
	} else {
		bucketDuration = int64(0)
	}

	return end, start, bucketDuration
}

func parseTimespan(r *http.Request) (int64, int64, int64) {
	var e string
	var s string
	var b string

	if v, ok := r.Form["end"]; ok && len(v) > 0 {
		e = v[0]
	}

	if v, ok := r.Form["start"]; ok && len(v) > 0 {
		s = v[0]
	}

	if v, ok := r.Form["bucketDuration"]; ok && len(v) > 0 {
		b = v[0]
	}

	return parseTimespanStrings(e, s, b)
}

func badID(w http.ResponseWriter, v bool) {
	w.WriteHeader(504)
	fmt.Fprintf(w, "{\"error\":\"504\",\"message\":\"Bad metrics IDe - 504\"}")

	if v {
		log.Printf("Bad metrics ID - 504\n")
	}
}

func baseTime(t int) int64 {
	nowSec := int64(time.Now().UTC().Unix())

	// if time is negative, assume base is now
	if t < 0 {
		return nowSec + int64(t)
	}

	return int64(t)
}

func parseSec(t string) (int64, error) {
	var err error
	var i int
	outSec := int64(0)

	// check for simple int
	if i, err = strconv.Atoi(t); err == nil {
		return baseTime(i / 1e3), nil
	} else if len(t) < 2 {
		// len < 2 and not int is an error
		return 0, err
	}

	// check for ms and mn
	switch t[len(t)-2:] {
	case "ms":
		if i, err = strconv.Atoi(t[:len(t)-2]); err == nil {
			return baseTime(i / 1e3), nil
		}
	case "mn":
		if i, err = strconv.Atoi(t[:len(t)-2]); err == nil {
			return baseTime(i * 60), nil
		}
	}

	// check for s, h and d
	switch t[len(t)-1:] {
	case "s":
		// if this is the "s" from "ms" then outSec != nil
		if i, err = strconv.Atoi(t[:len(t)-1]); err == nil && outSec == 0 {
			return baseTime(i), nil
		}
	case "h":
		if i, err = strconv.Atoi(t[:len(t)-1]); err == nil {
			return baseTime(i * 60 * 60), nil
		}
	case "d":
		if i, err = strconv.Atoi(t[:len(t)-1]); err == nil {
			return baseTime(i * 60 * 60 * 24), nil
		}
	}

	// if here must be an error
	return 0, err
}
