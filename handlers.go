package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/yaacov/mohawk/backends"
)

func ParseTags(tags string) map[string]string {
	vsf := map[string]string{}
	tagsList := strings.Split(tags, ",")
	for _, tag := range tagsList {
		t := strings.Split(tag, ":")
		if len(t) == 2 {
			vsf[t[0]] = t[1]
		}
	}
	return vsf
}

func (h Router) handleBadRequest(w http.ResponseWriter, r *http.Request) {
	var u interface{}
	json.NewDecoder(r.Body).Decode(&u)

	w.WriteHeader(404)

	fmt.Fprintf(w, "Error:")
	fmt.Fprintf(w, "Request: %+v\n", r)
	fmt.Fprintf(w, "Body: %+v\n", u)
}

func (h Router) handleStatus(w http.ResponseWriter, r *http.Request) {
	res := fmt.Sprintf("{\"MetricsService\":\"STARTED\",\"Implementation-Version\":\"%s\"}", VER)

	w.WriteHeader(200)
	fmt.Fprintln(w, res)
}

func (h Router) handleList(w http.ResponseWriter, r *http.Request) {

	res := []backend.Item{}
	
	if tags, ok := r.Form["tags"]; ok && len(tags) > 0 {
		fmt.Printf("%v\n\n", ParseTags(tags[0]))
		res = h.backend.GetItemList(ParseTags(tags[0]))
	} else {
		res = h.backend.GetItemList(map[string]string{})
	}
	resJson, _ := json.Marshal(res)

	w.WriteHeader(200)
	fmt.Fprintln(w, string(resJson))
}

func (h Router) handleGetData(w http.ResponseWriter, r *http.Request) {
	id := "id"
	end := int64(0)
	start := int64(time.Now().Unix() * 1000)
	limit := int64(100)
	order := "Asc"

	res := h.backend.GetRawData(id, end, start, limit, order)
	resJson, _ := json.Marshal(res)

	w.WriteHeader(200)
	fmt.Fprintln(w, string(resJson))
}

func (h Router) handlePushData(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	fmt.Fprintln(w, "{}")
}

func (h Router) handleUpdateTags(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	fmt.Fprintln(w, "{}")
}
