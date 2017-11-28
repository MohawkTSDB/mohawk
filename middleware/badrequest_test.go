package middleware

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBadRequestServeHTTP(t *testing.T) {
	a := BadRequestHandler(log.Printf)
	var (
		req *http.Request
		err error
	)

	testcases := []struct {
		method string
		status int
	}{
		{http.MethodGet, http.StatusNotFound},
		{http.MethodHead, http.StatusNotFound},
		{http.MethodOptions, http.StatusOK},
	}
	for _, tc := range testcases {
		if req, err = http.NewRequest(tc.method, "/", nil); err != nil {
			t.Error(err)
		}
		res := httptest.NewRecorder()
		a.ServeHTTP(res, req)

		if res.Code != tc.status {
			t.Errorf("expected status code to be %d but got %d", tc.status, res.Code)
		}
	}
}
