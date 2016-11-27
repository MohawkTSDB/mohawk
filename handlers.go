package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/yaacov/mohawk/backends"
)

type Handler struct {
	version string
	backend backend.Backend
}

func (h Handler) parseTags(tags string) map[string]string {
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

func (h Handler) BadRequest(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	var u interface{}
	json.NewDecoder(r.Body).Decode(&u)
	r.ParseForm()

	w.WriteHeader(404)

	fmt.Fprintf(w, "Error:")
	fmt.Fprintf(w, "Request: %+v\n", r)
	fmt.Fprintf(w, "Body: %+v\n", u)
}

func (h Handler) GetStatus(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	res := fmt.Sprintf("{\"MetricsService\":\"STARTED\",\"Implementation-Version\":\"%s\"}", h.version)

	w.WriteHeader(200)
	fmt.Fprintln(w, res)
}

func (h Handler) GetMetrics(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	res := []backend.Item{}
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
	r.ParseForm()

	id := argv["id"]
	end := int64(0)
	start := int64(time.Now().Unix() * 1000)
	limit := int64(100)
	order := "ASC"

	res := h.backend.GetRawData(id, end, start, limit, order)
	resJSON, _ := json.Marshal(res)

	w.WriteHeader(200)
	fmt.Fprintln(w, string(resJSON))
}
