package middleware

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestDefaultHeadersServeHTTP(t *testing.T) {
	a := DefaultHeadersDecorator()(func(w http.ResponseWriter, r *http.Request) {
	})
	var (
		req *http.Request
		err error
	)
	if req, err = http.NewRequest(http.MethodGet, "/something", nil); err != nil {
		t.Error(err)
	}
	res := httptest.NewRecorder()
	a.ServeHTTP(res, req)

	expected := http.Header{}
	expected.Set("Content-Type", "application/json")
	expected.Set("Access-Control-Allow-Origin", "*")
	expected.Set("Access-Control-Allow-Headers", "authorization,content-type,hawkular-tenant")
	expected.Set("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT")

	if !reflect.DeepEqual(expected, res.Header()) {
		t.Errorf("expected header to be equal to '%v' but got '%v'", expected, res.Header())
	}
}

func TestHeadersServeHTTP(t *testing.T) {
	okBody := "OK"
	a := HeadersDecorator(map[string]string{
		"User-Agent": "blabla",
	})(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, okBody)
	})
	var (
		req *http.Request
		err error
	)
	if req, err = http.NewRequest(http.MethodGet, "/something", nil); err != nil {
		t.Error(err)
	}
	res := httptest.NewRecorder()
	a.ServeHTTP(res, req)

	expected := http.Header{}
	expected.Set("User-Agent", "blabla")
	expected.Set("Content-Type", "text/plain; charset=utf-8") // this header is added once we are writing something to the body

	if !reflect.DeepEqual(expected, res.Header()) {
		t.Errorf("expected header to be equal to '%v' but got '%v'", expected, res.Header())
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}
	if string(body) != "OK" {
		t.Errorf("expected body '%s' to contain '%s'", string(body), "OK")
	}

}
