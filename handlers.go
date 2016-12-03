package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/yaacov/mohawk/backends"
)

// Handler common variables to be used by all Handler functions
// 	version the version of the Hawkular server we are mocking
// 	backend the backend to be used by the Handler functions
type Handler struct {
	version string
	backend backend.Backend
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

// GetStatus return a json status struct
func (h Handler) GetStatus(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	resTemplate := `{
	"MetricsService":"STARTED",
	"Implementation-Version":"%s",
	"MohawkVersion":"%s",
	"MohawkBackend":"%s"
}`
	res := fmt.Sprintf(resTemplate, h.version, VER, h.backend.Name())

	w.WriteHeader(200)
	fmt.Fprintln(w, res)
}

// GetAPIVersions return a json apiVersion struct
func (h Handler) GetAPIVersions(w http.ResponseWriter, r *http.Request, argv map[string]string) {
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

// GetMetrics return a list of metrics definitions
func (h Handler) GetMetrics(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	var res []backend.Item

	r.ParseForm()
	if tags, ok := r.Form["tags"]; ok && len(tags) > 0 {
		res = h.backend.GetItemList(parseTags(tags[0]))
	} else {
		res = h.backend.GetItemList(map[string]string{})
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
	} else {
		limit = int64(100)
	}
	if v, ok := r.Form["order"]; ok && len(v) > 0 {
		order = v[0]
		// do sanity check
		if order != "ASC" || order != "DESC" {
			order = "DESC"
		}
	} else {
		order = "DESC"
	}
	if v, ok := r.Form["bucketDuration"]; ok && len(v) > 0 {
		i, _ := strconv.Atoi(v[0][:len(v[0])-1])
		bucketDuration = int64(i)
	} else {
		bucketDuration = int64(0)
	}

	// call backend for data
	if bucketDuration == 0 {
		res := h.backend.GetRawData(id, end, start, limit, order)
		resJSON, _ := json.Marshal(res)
		resStr = string(resJSON)
	} else {
		res := h.backend.GetStatData(id, end, start, limit, order, bucketDuration)
		resJSON, _ := json.Marshal(res)
		resStr = string(resJSON)
	}

	// output to client
	w.WriteHeader(200)
	fmt.Fprintln(w, resStr)
}

// PostData send timestamp, value to the backend
func (h Handler) PostData(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	var u []map[string]interface{}
	json.NewDecoder(r.Body).Decode(&u)

	id := u[0]["id"].(string)
	t := u[0]["data"].([]interface{})[0].(map[string]interface{})["timestamp"].(float64)
	vStr := u[0]["data"].([]interface{})[0].(map[string]interface{})["value"].(string)
	v, _ := strconv.ParseFloat(vStr, 64)

	h.backend.PostRawData(id, int64(t), v)
	w.WriteHeader(200)
	fmt.Fprintln(w, "{}")
}

// PutTags send tag, value pairs to the backend
func (h Handler) PutTags(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	var tags map[string]string
	json.NewDecoder(r.Body).Decode(&tags)

	// use the id from the argv list
	id := argv["id"]

	h.backend.PutTags(id, tags)
	w.WriteHeader(200)
	fmt.Fprintln(w, "{}")
}
