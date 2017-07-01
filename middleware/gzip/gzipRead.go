package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
)

type GzBody struct {
	*gzip.Reader
	io.ReadCloser
}

func (r GzBody) Read(p []byte) (n int, err error) {
	return r.Reader.Read(p)
}

func (r GzBody) Close() error {
	return r.ReadCloser.Close()
}

func ReqBody(r *http.Request) io.ReadCloser {
	zr, err := gzip.NewReader(r.Body)
	if err != nil {
		return r.Body
	}

	return GzBody{zr, r.Body}
}
