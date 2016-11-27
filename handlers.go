package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/yaacov/mohawk/backends"
)

const VER = "0.21.0"

type Handler struct {
	backend backend.Backend
}

func (_ Handler) ParseTags(tags string) map[string]string {
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

func (h Handler) handleBadRequest(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	var u interface{}
	json.NewDecoder(r.Body).Decode(&u)

	w.WriteHeader(404)
	r.ParseForm()

	fmt.Fprintf(w, "Error:")
	fmt.Fprintf(w, "Request: %+v\n", r)
	fmt.Fprintf(w, "Body: %+v\n", u)
}

func (h Handler) handleStatus(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	res := fmt.Sprintf("{\"MetricsService\":\"STARTED\",\"Implementation-Version\":\"%s\"}", VER)

	w.WriteHeader(200)
	fmt.Fprintln(w, res)
}

func (h Handler) handleMetrics(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	res := []backend.Item{}

	if tags, ok := r.Form["tags"]; ok && len(tags) > 0 {
		res = h.backend.GetItemList(h.ParseTags(tags[0]))
	} else {
		res = h.backend.GetItemList(map[string]string{})
	}
	resJson, _ := json.Marshal(res)

	w.WriteHeader(200)
	fmt.Fprintln(w, string(resJson))
}

func (h Handler) handleGetData(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	r.ParseForm()

	id := argv["id"]
	end := int64(0)
	start := int64(time.Now().Unix() * 1000)
	limit := int64(100)
	order := "Asc"

	res := h.backend.GetRawData(id, end, start, limit, order)
	resJson, _ := json.Marshal(res)

	w.WriteHeader(200)
	fmt.Fprintln(w, string(resJson))
}

func (h Handler) handlePushData(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	w.WriteHeader(200)
	fmt.Fprintln(w, "{}")
}

func (h Handler) handleUpdateTags(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	w.WriteHeader(200)
	fmt.Fprintln(w, "{}")
}
