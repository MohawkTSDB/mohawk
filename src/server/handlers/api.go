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

// errBadMetricID a new error with bad metrics id message
var errBadMetricID = errors.New("Bad metrics ID")

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
func (h APIHhandler) GetAlertsStatus(w http.ResponseWriter, r *http.Request, argv map[string]string) error {
	if h.Alerts == nil {
		fmt.Fprintln(w, `{"AlertsService":"UNAVAILABLE"}`)
		return nil
	}

	resTemplate := `{"AlertsService":"STARTED","AlertsInterval":"%ds","Heartbeat":"%d","ServerURL":"%s"}`
	res := fmt.Sprintf(resTemplate, h.Alerts.AlertsInterval, h.Alerts.Heartbeat, h.Alerts.ServerURL)

	fmt.Fprintln(w, res)
	return nil
}

// GetAlerts return a list of alerts
func (h APIHhandler) GetAlerts(w http.ResponseWriter, r *http.Request, argv map[string]string) error {
	var res []alerts.Alert

	// if no alerts return empty list
	if h.Alerts == nil {
		fmt.Fprintln(w, `[]`)
		return nil
	}

	// get data from the form arguments
	if err := r.ParseForm(); err != nil {
		return err
	}

	// get tenant
	tenant := parseTenant(r)
	id := r.Form.Get("id")
	state := r.Form.Get("state")

	res = h.Alerts.FilterAlerts(tenant, id, state)
	resJSON, err := json.Marshal(res)
	if err != nil {
		return err
	}

	fmt.Fprintln(w, string(resJSON))
	return nil
}

// GetTenants return a list of metrics tenants
func (h APIHhandler) GetTenants(w http.ResponseWriter, r *http.Request, argv map[string]string) error {
	var res []storage.Tenant
	var resJSON []byte
	var err error

	if res, err = h.Storage.GetTenants(); err == nil {
		resJSON, err = json.Marshal(res)
		if err != nil {
			return err
		}
	}

	fmt.Fprintln(w, string(resJSON))
	return nil
}

// GetMetricsHelper helper function to return a list of metrics definitions
func (h APIHhandler) GetMetricsHelper(w http.ResponseWriter, r *http.Request, argv map[string]string) ([]storage.Item, error) {
	var res []storage.Item
	var err error
	var tags = map[string]string{}

	// get data from the form arguments
	if err = r.ParseForm(); err != nil {
		return res, err
	}

	// we only use gauges
	if typeStr, ok := r.Form["type"]; ok && len(typeStr) > 0 && typeStr[0] != "gauge" {
		return res, nil
	}

	// get tenant
	tenant := parseTenant(r)

	// get a list of gauges
	if tagsStr, ok := r.Form["tags"]; ok && len(tagsStr) > 0 {
		tags = storage.ParseTags(tagsStr[0])
		if !validTags(tags) {
			return res, errBadMetricID
		}
	}

	return h.Storage.GetItemList(tenant, tags)
}

// GetMetrics return a list of metrics definitions
func (h APIHhandler) GetMetrics(w http.ResponseWriter, r *http.Request, argv map[string]string) error {
	var res []storage.Item
	var resJSON []byte
	var err error

	res, err = h.GetMetricsHelper(w, r, argv)
	if err != nil {
		return err
	}

	if resJSON, err = json.Marshal(res); err == nil {
		fmt.Fprintln(w, string(resJSON))
	}

	return err
}

// GetExports return a list of metrics if prmetheus export format
func (h APIHhandler) GetExports(w http.ResponseWriter, r *http.Request, argv map[string]string) error {
	var res []storage.Item
	var err error
	var first bool

	res, err = h.GetMetricsHelper(w, r, argv)
	if err != nil {
		return err
	}

	for _, i := range res {
		// check for last values
		if len(i.LastValues) < 1 {
			continue
		}

		// print name
		if name, ok := i.Tags["__name__"]; ok {
			fmt.Fprintf(w, name)
		} else {
			fmt.Fprintf(w, i.ID)
		}

		// print tags
		first = true
		fmt.Fprintf(w, "{")
		for k, v := range i.Tags {
			// __name__ is not a real tag
			if k == "__name__" {
				continue
			}

			if first {
				fmt.Fprintf(w, "%s=\"%s\"", k, v)
				first = false
			} else {
				fmt.Fprintf(w, ",%s=\"%s\"", k, v)
			}
		}
		fmt.Fprintf(w, "} ")

		// print value (if we are here i.LastValues[0] should exist)
		fmt.Fprintf(w, "%f\n", i.LastValues[0].Value)
	}

	return err
}

// GetData return a list of metrics raw / stat data
func (h APIHhandler) GetData(w http.ResponseWriter, r *http.Request, argv map[string]string) error {
	// use the id from the argv list
	id := argv["id"]
	if !validStr(id) {
		return errBadMetricID
	}

	// get data from the form arguments
	if err := r.ParseForm(); err != nil {
		return err
	}

	// get tenant
	tenant := parseTenant(r)

	// get timespan
	end, start, bucketDuration, err := parseTimespan(r)
	if err != nil {
		return err
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
	// call storage for data
	return h.getData(w, tenant, id, end, start, limit, order, bucketDuration)
}

// DeleteData delete a list of metrics raw  data
func (h APIHhandler) DeleteData(w http.ResponseWriter, r *http.Request, argv map[string]string) error {
	// use the id from the argv list
	id := argv["id"]
	if !validStr(id) {
		return errBadMetricID
	}

	// get data from the form arguments
	if err := r.ParseForm(); err != nil {
		return err
	}

	// get tenant
	tenant := parseTenant(r)

	// get timespan
	end, start, _, err := parseTimespan(r)
	if err != nil {
		return err
	}

	if h.Verbose {
		log.Printf("ID: %s@%s, End: %d, Start: %d", tenant, id, end, start)
	}

	// call storage for data
	if start < end {
		err = h.Storage.DeleteData(tenant, id, end, start)

		// output to client
		if err == nil {
			fmt.Fprintf(w, "{\"message\":\"Deleted %s@%s [%d-%d]\"}", tenant, id, end, start)
		}

		return err
	}

	return errors.New("Can't delete time range")
}

// PostMQuery query data from storage + gauges
func (h APIHhandler) PostMQuery(w http.ResponseWriter, r *http.Request, argv map[string]string) error {
	// parse query args
	tenant, ids, end, start, limit, order, bucketDuration, err := h.parseQueryArgs(w, r, argv)
	if err != nil {
		return err
	}
	numOfItems := len(ids) - 1

	fmt.Fprintf(w, "{\"gauge\":{")

	for i, id := range ids {
		// write data
		fmt.Fprintf(w, "\"%s\":", id)

		// call storage for data, and send it to writer
		if err = h.getData(w, tenant, id, end, start, limit, order, bucketDuration); err != nil {
			return err
		}

		if i < numOfItems {
			fmt.Fprintf(w, ",")
		}
	}

	fmt.Fprintf(w, "}}")
	return nil
}

// PostQuery query data from storage
func (h APIHhandler) PostQuery(w http.ResponseWriter, r *http.Request, argv map[string]string) error {
	// parse query args
	tenant, ids, end, start, limit, order, bucketDuration, err := h.parseQueryArgs(w, r, argv)
	if err != nil {
		return err
	}
	numOfItems := len(ids) - 1

	fmt.Fprintf(w, "[")

	for i, id := range ids {
		// write data
		fmt.Fprintf(w, "{\"id\": \"%s\", \"data\":", id)

		// call storage for data, and send it to writer
		if err := h.getData(w, tenant, id, end, start, limit, order, bucketDuration); err != nil {
			return err
		}

		fmt.Fprintf(w, "}")

		if i < numOfItems {
			fmt.Fprintf(w, ",")
		}
	}

	fmt.Fprintf(w, "]")
	return nil
}

// PostData send timestamp, value to the storage
func (h APIHhandler) PostData(w http.ResponseWriter, r *http.Request, argv map[string]string) error {
	var u []postDataItems
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		return err
	}

	for _, item := range u {
		if !validStr(item.ID) {
			return errBadMetricID
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

			if err := h.Storage.PostRawData(tenant, id, timestamp, value); err != nil {
				return err
			}
		}
	}

	fmt.Fprintf(w, "{\"message\":\"Received %d data items\"}", len(u))
	return nil
}

// PutTags send tag, value pairs to the storage
func (h APIHhandler) PutTags(w http.ResponseWriter, r *http.Request, argv map[string]string) error {
	var tags map[string]string
	if err := json.NewDecoder(r.Body).Decode(&tags); err != nil {
		return err
	}

	// use the id from the argv list
	id := argv["id"]
	if !validStr(id) || !validTags(tags) {
		return errBadMetricID
	}

	// get tenant
	tenant := parseTenant(r)

	if h.Verbose {
		log.Printf("Tenant: %s, ID: %+v {tags: %+v}\n", tenant, id, tags)
	}

	if err := h.Storage.PutTags(tenant, id, tags); err != nil {
		return err
	}

	fmt.Fprintf(w, "{\"message\":\"Updated tags for %s@%s\"}", tenant, id)
	return nil
}

// PutMultiTags send tags pet dataItem - tag, value pairs to the storage
func (h APIHhandler) PutMultiTags(w http.ResponseWriter, r *http.Request, argv map[string]string) error {
	var u []putTags
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		return err
	}

	for _, item := range u {
		if !validStr(item.ID) {
			return errBadMetricID
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
			if err := h.Storage.PutTags(tenant, id, item.Tags); err != nil {
				return err
			}
		}
	}

	fmt.Fprintf(w, "{\"message\":\"Updated tags for %d items\"}", len(u))
	return nil
}

// DeleteTags delete a tag
func (h APIHhandler) DeleteTags(w http.ResponseWriter, r *http.Request, argv map[string]string) error {
	// use the id from the argv list
	id := argv["id"]
	tagsStr := argv["tags"]
	if !validStr(id) || !validStr(tagsStr) {
		return errBadMetricID
	}
	tags := strings.Split(tagsStr, ",")

	// get tenant
	tenant := parseTenant(r)

	if err := h.Storage.DeleteTags(tenant, id, tags); err != nil {
		return err
	}

	fmt.Fprintf(w, "{\"message\":\"Deleted tags for %s@%s\"}", tenant, id)
	return nil
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
			return tenant, u, errBadMetricID
		}
	}

	// add ids from tags query
	if u.Tags != "" {
		res, _ := h.Storage.GetItemList(tenant, storage.ParseTags(u.Tags))
		for _, r := range res {
			u.IDs = append(u.IDs, r.ID)
		}
	}

	return tenant, u, nil
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
func (h APIHhandler) getData(w http.ResponseWriter, tenant string, id string, end int64, start int64, limit int64, order string, bucketDuration int64) error {
	var resJSON []byte
	var err error

	// call storage for data
	if bucketDuration == 0 {
		if res, errQuery := h.Storage.GetRawData(tenant, id, end, start, limit, order); errQuery == nil {
			resJSON, err = json.Marshal(res)
		}
	} else {
		if res, errQuery := h.Storage.GetStatData(tenant, id, end, start, limit, order, bucketDuration); errQuery == nil {
			resJSON, err = json.Marshal(res)
		}
	}
	if err == nil {
		fmt.Fprintf(w, string(resJSON))
	}
	return err
}
