// Copyright 2016,2017,2018 Yaacov Zamir <kobi.zamir@gmail.com>
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
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/MohawkTSDB/mohawk/src/alerts"
	"github.com/MohawkTSDB/mohawk/src/storage"
)

// const defaultLimit default REST API call query limit
const defaultLimit = 20000

// const defaultOrder default REST API call query order
const defaultOrder = "ASC"

// const secondaryOrder secondary REST API call query order
const secondaryOrder = "DESC"

// APIHhandler common variables to be used by all APIHhandler functions
// 	version the version of the Hawkular server we are mocking
// 	storage the storage to be used by the APIHhandler functions
type APIHhandler struct {
	Verbose bool
	Storage storage.Storage
	Alerts  *alerts.AlertRules
}

// GetAlertsStatus return a json alerts status struct
func (h APIHhandler) GetAlertsStatus(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	if h.Alerts == nil {
		w.WriteHeader(200)
		fmt.Fprintln(w, `{"AlertsService":"UNAVAILABLE"}`)
		return
	}

	resTemplate := `{"AlertsService":"STARTED","AlertsInterval":"%ds","Heartbeat":"%d","ServerURL":"%s"}`
	res := fmt.Sprintf(resTemplate, h.Alerts.AlertsInterval, h.Alerts.Heartbeat, h.Alerts.ServerURL)

	w.WriteHeader(200)
	fmt.Fprintln(w, res)
}

// GetAlerts return a list of alerts
func (h APIHhandler) GetAlerts(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	var res []alerts.Alert

	// if no alerts return empty list
	if h.Alerts == nil {
		w.WriteHeader(200)
		fmt.Fprintln(w, `[]`)
		return
	}

	// get data from the form arguments
	// get data from the form arguments
	if err := r.ParseForm(); err != nil {
		if h.Verbose {
			log.Printf(err.Error())
		}
		w.WriteHeader(504)
		fmt.Fprintf(w, "{\"error\":\"504\",\"message\":\"%s\"}", err.Error())
		return
	}

	// get tenant
	tenant := parseTenant(r)
	id := r.Form.Get("id")
	state := r.Form.Get("state")

	res = h.Alerts.FilterAlerts(tenant, id, state)
	resJSON, _ := json.Marshal(res)

	w.WriteHeader(200)
	fmt.Fprintln(w, string(resJSON))
}

// GetTenants return a list of metrics tenants
func (h APIHhandler) GetTenants(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	var res []storage.Tenant

	res = h.Storage.GetTenants()
	resJSON, _ := json.Marshal(res)

	w.WriteHeader(200)
	fmt.Fprintln(w, string(resJSON))
}

// GetMetrics return a list of metrics definitions
func (h APIHhandler) GetMetrics(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	var res []storage.Item

	// get data from the form arguments
	if err := r.ParseForm(); err != nil {
		if h.Verbose {
			log.Printf(err.Error())
		}
		w.WriteHeader(504)
		fmt.Fprintf(w, "{\"error\":\"504\",\"message\":\"%s\"}", err.Error())
		return
	}

	// we only use gauges
	if typeStr, ok := r.Form["type"]; ok && len(typeStr) > 0 && typeStr[0] != "gauge" {
		w.WriteHeader(200)
		fmt.Fprintln(w, "[]")
		return
	}

	// get tenant
	tenant := parseTenant(r)

	// get a list of gauges
	if tagsStr, ok := r.Form["tags"]; ok && len(tagsStr) > 0 {
		tags := storage.ParseTags(tagsStr[0])
		if !validTags(tags) {
			badID(w, h.Verbose)
			return
		}
		res = h.Storage.GetItemList(tenant, tags)
	} else {
		res = h.Storage.GetItemList(tenant, map[string]string{})
	}
	resJSON, _ := json.Marshal(res)

	w.WriteHeader(200)
	fmt.Fprintln(w, string(resJSON))
}

// GetData return a list of metrics raw / stat data
func (h APIHhandler) GetData(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	// use the id from the argv list
	id := argv["id"]
	if !validStr(id) {
		badID(w, h.Verbose)
		return
	}

	// get data from the form arguments
	if err := r.ParseForm(); err != nil {
		if h.Verbose {
			log.Printf(err.Error())
		}
		w.WriteHeader(504)
		fmt.Fprintf(w, "{\"error\":\"504\",\"message\":\"%s\"}", err.Error())
		return
	}

	// get tenant
	tenant := parseTenant(r)

	// get timespan
	end, start, bucketDuration, err := parseTimespan(r)
	if err != nil {
		if h.Verbose {
			log.Printf(err.Error())
		}
		w.WriteHeader(504)
		fmt.Fprintf(w, "{\"error\":\"504\",\"message\":\"%s\"}", err.Error())
		return
	}

	limit := int64(defaultLimit)
	if v, ok := r.Form["limit"]; ok && len(v) > 0 {
		if i, err := strconv.Atoi(v[0]); err == nil && i > 0 {
			limit = int64(i)
		}
	}

	order := defaultOrder
	if v, ok := r.Form["order"]; ok && len(v) > 0 && v[0] == secondaryOrder {
		order = secondaryOrder
	}

	if h.Verbose {
		log.Printf("ID: %s@%s, End: %d, Start: %d, Limit: %d, Order: %s, bucketDuration: %ds", tenant, id, end, start, limit, order, bucketDuration)
	}

	// output to client
	w.WriteHeader(200)

	// call storage for data
	h.getData(w, tenant, id, end, start, limit, order, bucketDuration)
}

// DeleteData delete a list of metrics raw  data
func (h APIHhandler) DeleteData(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	// use the id from the argv list
	id := argv["id"]
	if !validStr(id) {
		badID(w, h.Verbose)
		return
	}

	// get data from the form arguments
	if err := r.ParseForm(); err != nil {
		if h.Verbose {
			log.Printf(err.Error())
		}
		w.WriteHeader(504)
		fmt.Fprintf(w, "{\"error\":\"504\",\"message\":\"%s\"}", err.Error())
		return
	}

	// get tenant
	tenant := parseTenant(r)

	// get timespan
	end, start, _, err := parseTimespan(r)
	if err != nil {
		if h.Verbose {
			log.Printf(err.Error())
		}
		w.WriteHeader(504)
		fmt.Fprintf(w, "{\"error\":\"504\",\"message\":\"%s\"}", err.Error())
		return
	}

	if h.Verbose {
		log.Printf("ID: %s@%s, End: %d, Start: %d", tenant, id, end, start)
	}

	// call storage for data
	if start < end {
		h.Storage.DeleteData(tenant, id, end, start)

		// output to client
		w.WriteHeader(200)
		fmt.Fprintf(w, "{\"message\":\"Deleted %s@%s [%d-%d]\"}", tenant, id, end, start)
		return
	}

	w.WriteHeader(504)
	fmt.Fprintf(w, "{\"error\":\"504\",\"message\":\"Can't delete time rage - 504\"}")
}

// PostMQuery query data from storage + gauges
func (h APIHhandler) PostMQuery(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	// parse query args
	tenant, ids, end, start, limit, order, bucketDuration, err := h.parseQueryArgs(w, r, argv)
	if err != nil {
		if h.Verbose {
			log.Printf(err.Error())
		}
		w.WriteHeader(504)
		fmt.Fprintf(w, "{\"error\":\"504\",\"message\":\"%s\"}", err.Error())
		return
	}
	numOfItems := len(ids) - 1

	w.WriteHeader(200)
	fmt.Fprintf(w, "{\"gauge\":{")

	for i, id := range ids {
		// write data
		fmt.Fprintf(w, "\"%s\":", id)

		// call storage for data, and send it to writer
		h.getData(w, tenant, id, end, start, limit, order, bucketDuration)

		if i < numOfItems {
			fmt.Fprintf(w, ",")
		}
	}

	fmt.Fprintf(w, "}}")
}

// PostQuery query data from storage
func (h APIHhandler) PostQuery(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	// parse query args
	tenant, ids, end, start, limit, order, bucketDuration, err := h.parseQueryArgs(w, r, argv)
	if err != nil {
		if h.Verbose {
			log.Printf(err.Error())
		}
		w.WriteHeader(504)
		fmt.Fprintf(w, "{\"error\":\"504\",\"message\":\"%s\"}", err.Error())
		return
	}
	numOfItems := len(ids) - 1

	w.WriteHeader(200)
	fmt.Fprintf(w, "[")

	for i, id := range ids {
		// write data
		fmt.Fprintf(w, "{\"id\": \"%s\", \"data\":", id)

		// call storage for data, and send it to writer
		h.getData(w, tenant, id, end, start, limit, order, bucketDuration)

		fmt.Fprintf(w, "}")

		if i < numOfItems {
			fmt.Fprintf(w, ",")
		}
	}

	fmt.Fprintf(w, "]")
}

// PostData send timestamp, value to the storage
func (h APIHhandler) PostData(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	var u []postDataItems
	json.NewDecoder(r.Body).Decode(&u)

	for _, item := range u {
		if !validStr(item.ID) {
			badID(w, h.Verbose)
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

			if h.Verbose {
				log.Printf("Tenant: %s, ID: %+v {timestamp: %+v, value: %+v}\n", tenant, id, timestamp, value)
			}
			h.Storage.PostRawData(tenant, id, timestamp, value)
		}
	}

	w.WriteHeader(200)
	fmt.Fprintf(w, "{\"message\":\"Received %d data items\"}", len(u))
}

// PutTags send tag, value pairs to the storage
func (h APIHhandler) PutTags(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	var tags map[string]string
	json.NewDecoder(r.Body).Decode(&tags)

	// use the id from the argv list
	id := argv["id"]
	if !validStr(id) || !validTags(tags) {
		badID(w, h.Verbose)
		return
	}

	// get tenant
	tenant := parseTenant(r)

	if h.Verbose {
		log.Printf("Tenant: %s, ID: %+v {tags: %+v}\n", tenant, id, tags)
	}
	h.Storage.PutTags(tenant, id, tags)

	w.WriteHeader(200)
	fmt.Fprintf(w, "{\"message\":\"Updated tags for %s@%s\"}", tenant, id)
}

// PutMultiTags send tags pet dataItem - tag, value pairs to the storage
func (h APIHhandler) PutMultiTags(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	var u []putTags
	json.NewDecoder(r.Body).Decode(&u)

	for _, item := range u {
		if !validStr(item.ID) {
			badID(w, h.Verbose)
			return
		}
	}

	// get tenant
	tenant := parseTenant(r)

	for _, item := range u {
		id := item.ID
		if validTags(item.Tags) {
			if h.Verbose {
				log.Printf("Tenant: %s, ID: %+v {tags: %+v}\n", tenant, id, item.Tags)
			}
			h.Storage.PutTags(tenant, id, item.Tags)
		}
	}

	w.WriteHeader(200)
	fmt.Fprintf(w, "{\"message\":\"Updated tags for %d items\"}", len(u))
}

// DeleteTags delete a tag
func (h APIHhandler) DeleteTags(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	// use the id from the argv list
	id := argv["id"]
	tagsStr := argv["tags"]
	if !validStr(id) || !validStr(tagsStr) {
		badID(w, h.Verbose)
		return
	}
	tags := strings.Split(tagsStr, ",")

	// get tenant
	tenant := parseTenant(r)

	h.Storage.DeleteTags(tenant, id, tags)

	w.WriteHeader(200)
	fmt.Fprintf(w, "{\"message\":\"Deleted tags for %s@%s\"}", tenant, id)
}

// decodeRequestBody parse request body
func (h APIHhandler) decodeRequestBody(r *http.Request) (tenant string, u dataQuery, err error) {
	// get tenant
	tenant = parseTenant(r)

	// decode query body
	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()
	if err = decoder.Decode(&u); err != nil {
		return tenant, u, err
	}

	// get ids from explicit ids list
	for _, id := range u.IDs {
		if !validStr(id) {
			err = errors.New("Bad metrics ID - 504")
			return tenant, u, err
		}
	}

	// add ids from tags query
	if u.Tags != "" {
		res := h.Storage.GetItemList(tenant, storage.ParseTags(u.Tags))
		for _, r := range res {
			u.IDs = append(u.IDs, r.ID)
		}
	}

	return tenant, u, err
}

// parseQueryArgs parse query request body args
func (h APIHhandler) parseQueryArgs(w http.ResponseWriter, r *http.Request, argv map[string]string) (tenant string, ids []string, end int64, start int64, limit int64, order string, bucketDuration int64, err error) {
	var endStr string
	var startStr string
	var bucketDurationStr string

	// get tenant
	tenant, u, err := h.decodeRequestBody(r)
	if err != nil {
		return tenant, []string{}, 0, 0, 0, "", 0, err
	}

	// get start time string
	switch v := u.Start.(type) {
	case string:
		startStr = u.Start.(string)
	case nil:
		startStr = ""
	default:
		startStr = fmt.Sprintf("%+v", v)
	}

	// get end time string
	switch v := u.End.(type) {
	case string:
		endStr = u.End.(string)
	case nil:
		endStr = ""
	default:
		endStr = fmt.Sprintf("%+v", v)
	}

	// get bucket duration string
	switch v := u.BucketDuration.(type) {
	case string:
		bucketDurationStr = u.BucketDuration.(string)
	case nil:
		bucketDurationStr = ""
	default:
		bucketDurationStr = fmt.Sprintf("%+v", v)
	}

	// get query items limit
	if limit, err = u.Limit.Int64(); err != nil || limit < 1 {
		// using default value, remove error
		limit = int64(defaultLimit)
		err = nil
	}

	// get query order
	order = defaultOrder
	if u.Order == secondaryOrder {
		order = secondaryOrder
	}

	// calc timestamps from end, start and bucket duration strings
	end, start, bucketDuration, err = parseTimespanStrings(endStr, startStr, bucketDurationStr)

	if h.Verbose {
		log.Printf("Tenant: %s, IDs: %+v", tenant, u.IDs)
		log.Printf("End: %d(%s), Start: %d(%s), Limit: %d, Order: %s, bucketDuration: %ds", end, endStr, start, startStr, limit, order, bucketDuration)
	}

	return tenant, u.IDs, end, start, limit, order, bucketDuration, err
}

// getData querys data from the storage, and send it to writer
func (h APIHhandler) getData(w http.ResponseWriter, tenant string, id string, end int64, start int64, limit int64, order string, bucketDuration int64) {
	var resStr string

	// call storage for data
	if bucketDuration == 0 {
		res := h.Storage.GetRawData(tenant, id, end, start, limit, order)
		resJSON, _ := json.Marshal(res)
		resStr = string(resJSON)
	} else {
		res := h.Storage.GetStatData(tenant, id, end, start, limit, order, bucketDuration)
		resJSON, _ := json.Marshal(res)
		resStr = string(resJSON)
	}

	fmt.Fprintf(w, resStr)
}
