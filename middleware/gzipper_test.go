package middleware

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"compress/gzip"
)

func TestGzipServeHTTP(t *testing.T) {
	a := GzipDecodeDecorator()(GzipEncodeDecorator()(func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			t.Error("body was not supposed to be nil")
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		fmt.Fprintf(w, string(body))
	}))
	var (
		req  *http.Request
		err  error
		body []byte
	)

	testcases := []struct {
		useGzip  bool
		header   map[string]string
		bodyStr  string
		expected string
		toGunzip bool
		toGzip   bool
	}{
		{true, map[string]string{}, "none", "none", false, false},
		{false, map[string]string{}, "none", "none", false, false},
		{true, map[string]string{"Accept-Encoding": "gzip"}, "bla", "bla", true, false},
		{true, map[string]string{"Content-Encoding": "gzip"}, "gzbla", "gzbla", false, true},
		{true, map[string]string{"Content-Encoding": "gzip", "Accept-Encoding": "gzip"}, "gzbla", "gzbla", true, true},
	}
	var reader io.Reader
	for _, tc := range testcases {
		//a.UseGzip = tc.useGzip
		reader = bytes.NewBufferString(tc.bodyStr)
		if tc.toGzip {
			buff := bytes.NewBuffer(nil)
			gWriter := gzip.NewWriter(buff)
			gWriter.Write([]byte(tc.bodyStr))
			if err := gWriter.Flush(); err != nil {
				t.Error(err)
			}
			if err := gWriter.Close(); err != nil {
				t.Error(err)
			}
			reader = buff
		}
		if req, err = http.NewRequest(http.MethodGet, "/", reader); err != nil {
			t.Error(err)
		}
		if len(tc.header) > 0 {
			for k, v := range tc.header {
				req.Header.Set(k, v)
			}
		}
		res := httptest.NewRecorder()
		a.ServeHTTP(res, req)

		var bodyReader io.Reader
		bodyReader = res.Body
		if tc.toGunzip {
			bodyReader, err = gzip.NewReader(bodyReader)
			if err != nil {
				t.Error(err)
			}
		}
		if body, err = ioutil.ReadAll(bodyReader); err != nil {
			t.Error(err)
		}

		if string(body) != tc.expected {
			t.Errorf("expected response body to be '%s' but got '%s'", tc.expected, string(body))
		}
	}
}
