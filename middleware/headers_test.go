package middleware

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestHeadersServeHTTP(t *testing.T) {
	a := Headers{
		next: HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		}),
	}
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
