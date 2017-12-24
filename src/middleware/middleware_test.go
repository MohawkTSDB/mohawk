package middleware

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/MohawkTSDB/mohawk/src/router"
)

func TestAppendMiddleware(t *testing.T) {
	myRoute := router.Router{
		Prefix: "/hawkular/bla/",
	}
	myRoute.Add("GET", ":id/info", func(w http.ResponseWriter, r *http.Request, argv map[string]string) {
		fmt.Fprintf(w, "{\"msg\":\"got data\"")
	})

}
