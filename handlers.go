package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/yaacov/mohawk/backends"
)

type Handler struct {
	version string
	backend backend.Backend
}

func (h Handler) parseTags(tags string) map[string]string {
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

func (h Handler) BadRequest(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	var u interface{}

	json.NewDecoder(r.Body).Decode(&u)
	r.ParseForm()

	w.WriteHeader(404)

	log.Printf("BadRequest:\n")
	log.Printf("Request: %+v\n", r)
	log.Printf("Body: %+v\n", u)

	fmt.Fprintf(w, "Error:")
	fmt.Fprintf(w, "Request: %+v\n", r)
	fmt.Fprintf(w, "Body: %+v\n", u)
}

func (h Handler) GetStatus(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	res := fmt.Sprintf(`{
		"MetricsService":"STARTED",
		"Implementation-Version":"%s",
		"MoHawk":"%s"
		}`, h.version, VER)

	w.WriteHeader(200)
	fmt.Fprintln(w, res)
}

func (h Handler) GetAPIVersions(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	w.WriteHeader(200)
	fmt.Fprintln(w, `{
	  "kind": "APIVersions",
	  "apiVersion": "v1",
	  "versions": [
	    "v1"
	  ],
	  "serverAddressByClientCIDRs": null
	}`)
}

func (h Handler) GetMetrics(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	var res []backend.Item

	r.ParseForm()
	if tags, ok := r.Form["tags"]; ok && len(tags) > 0 {
		res = h.backend.GetItemList(h.parseTags(tags[0]))
	} else {
		res = h.backend.GetItemList(map[string]string{})
	}
	resJSON, _ := json.Marshal(res)

	w.WriteHeader(200)
	fmt.Fprintln(w, string(resJSON))
}

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
		start = end - int64(8 * 60 * 60 * 1000)
	}
	if v, ok := r.Form["limit"]; ok && len(v) > 0 {
		i, _ := strconv.Atoi(v[0])
		limit = int64(i)
	} else {
		limit = int64(100)
	}
	if v, ok := r.Form["order"]; ok && len(v) > 0 {
		order = v[0]
	} else {
		order = "DESC"
	}
	if v, ok := r.Form["bucketDuration"]; ok && len(v) > 0 {
		i, _ := strconv.Atoi(v[0][:len(v[0]) - 1])
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

func (h Handler) PutTags(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	var tags map[string]string

	json.NewDecoder(r.Body).Decode(&tags)

	// use the id from the argv list
	id := argv["id"]

	h.backend.PutTags(id, tags)

	w.WriteHeader(200)
	fmt.Fprintln(w, "{}")
}
