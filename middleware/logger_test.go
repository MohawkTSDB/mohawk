package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggerServeHTTP(t *testing.T) {
	a := Logger{
		next: HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "OK")
		}),
	}
	var (
		req *http.Request
		err error
	)

	buff := bytes.NewBufferString("")
	testcases := []struct {
		logFunc logFunc
	}{
		{nil},
		{logFunc(func(format string, v ...interface{}) {
			fmt.Fprintf(buff, format, v...)
		})},
	}
	for _, tc := range testcases {
		buff.Reset()
		a.logFunc = tc.logFunc
		if req, err = http.NewRequest(http.MethodGet, "/", nil); err != nil {
			t.Error(err)
		}
		res := httptest.NewRecorder()
		a.ServeHTTP(res, req)

		if tc.logFunc != nil {
			expected := fmt.Sprintf("%s Accept-Encoding: %s, %4s %s", req.RemoteAddr, req.Header.Get("Accept-Encoding"), req.Method, req.URL)
			if buff.String() != expected {
				t.Errorf("expected output to be '%s' but got '%s'", expected, buff)
			}
		}
	}
}
