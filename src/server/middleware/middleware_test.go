package middleware

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/MohawkTSDB/mohawk/src/server/router"
)

func TestAppendMiddleware(t *testing.T) {
	myRoute := router.Router{
		Prefix: "/hawkular/bla/",
	}
	myRoute.Add("GET", ":id/info", func(w http.ResponseWriter, r *http.Request, argv map[string]string) error {
		fmt.Fprintf(w, "{\"msg\":\"got data\"")
		return nil
	})

}
