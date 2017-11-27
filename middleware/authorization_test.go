package middleware

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthServeHTTP(t *testing.T) {
	OkBody := "DONE"
	a := Authorization{
		UseToken:        true,
		PublicPathRegex: "^/bla/boy",
		Token:           "ninja",
		next: HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, OkBody)
		}),
	}

	testcases := []struct {
		path    string
		token   string
		code    int
		content string
	}{
		{"/bla/boy/1", "", http.StatusOK, OkBody},
		{"/hello/world", "", http.StatusUnauthorized, "Unauthorized - 401"},
		{"/hello/world", a.Token, http.StatusOK, OkBody},
		{"/bla/boy/2", a.Token, http.StatusOK, OkBody},
	}

	var (
		err  error
		req  *http.Request
		body []byte
	)

	for _, tc := range testcases {
		if req, err = http.NewRequest(http.MethodGet, tc.path, nil); err != nil {
			t.Error(err)
		}
		if tc.token != "" {
			req.Header.Set("Authorization", "Bearer "+tc.token)
		}
		res := httptest.NewRecorder()
		a.ServeHTTP(res, req)

		if body, err = ioutil.ReadAll(res.Body); err != nil {
			t.Error(err)
		}
		if !strings.Contains(string(body), tc.content) {
			t.Errorf("expected body '%s' to contain '%s' path %s", string(body), tc.content, tc.path)
		}
		if res.Code != tc.code {
			t.Errorf("expected status code to be %d but got %d path %s", tc.code, res.Code, tc.path)
		}
	}

}
