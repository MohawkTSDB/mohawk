package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggerServeHTTP(t *testing.T) {
	var (
		req *http.Request
		err error
	)

	buff := bytes.NewBufferString("")
	testcases := []struct {
		logFunc func(string, ...interface{})
	}{
		{func(format string, v ...interface{}) {
			fmt.Fprintf(buff, format, v...)
		}},
	}
	for _, tc := range testcases {
		buff.Reset()
		a := LoggingDecorator(tc.logFunc)(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "OK")
		})
		if req, err = http.NewRequest(http.MethodGet, "/", nil); err != nil {
			t.Error(err)
		}
		res := httptest.NewRecorder()
		a.ServeHTTP(res, req)

		expected := fmt.Sprintf("%s Accept-Encoding: %s, %4s %s", req.RemoteAddr, req.Header.Get("Accept-Encoding"), req.Method, req.URL)
		if buff.String() != expected {
			t.Errorf("expected output to be '%s' but got '%s'", expected, buff)
		}
	}
}
