package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
)

type ResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w ResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func Decorator(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")

		gz := gzip.NewWriter(w)
		defer gz.Close()

		responseWriter := ResponseWriter{Writer: gz, ResponseWriter: w}
		handler(responseWriter, r)
	}
}
