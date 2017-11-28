package middleware

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/MohawkTSDB/mohawk/router"
)

func TestAppendMiddleware(t *testing.T) {
	myRoute := router.Router{
		Prefix: "/hawkular/bla/",
	}
	myRoute.Add("GET", ":id/info", func(w http.ResponseWriter, r *http.Request, argv map[string]string) {
		fmt.Fprintf(w, "got data")
	})

}
