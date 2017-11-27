package middleware

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStaticServeHTTP(t *testing.T) {
	a := Static{
		MediaPath: "./tests",
		next: HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "OK")
		}),
	}
	var (
		req  *http.Request
		err  error
		body []byte
	)

	testcases := []struct {
		reqURL  string
		expBody string
	}{
		{"/", "<div>Hello world</div>"},
		{"/index.html", "<div>Hello world</div>"},
		{"/inner", "<div>inner hello</div>"},
		{"/index1.html", "OK"},
	}
	for _, tc := range testcases {
		if req, err = http.NewRequest(http.MethodGet, tc.reqURL, nil); err != nil {
			t.Error(err)
		}
		res := httptest.NewRecorder()
		a.ServeHTTP(res, req)

		if body, err = ioutil.ReadAll(res.Body); err != nil {
			t.Error(err)
		}

		if string(body) != tc.expBody {
			t.Errorf("expected response body to be equal to '%s' but got '%s'", tc.expBody, string(body))
		}
	}
}
