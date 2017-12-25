// Copyright 2016,2017 Yaacov Zamir <kobi.zamir@gmail.com>
// and other contributors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package middleware middlewares for Mohawk
package middleware

import (
	"io"
	"net/http"
	"strings"

	"compress/gzip"
)

type gzBody struct {
	*gzip.Reader
	io.ReadCloser
}

func (r gzBody) Read(p []byte) (n int, err error) {
	return r.Reader.Read(p)
}

func (r gzBody) Close() error {
	return r.ReadCloser.Close()
}

type gzRespWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzRespWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func GzipDecodeDecorator() Decorator {
	return Decorator(func(h http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Content-Encoding") == "gzip" {
				// TODO: I personally think it's a good idea to return
				// 		 http.StatusBadRequest if err != nil
				gzReader, err := gzip.NewReader(r.Body)
				if err == nil {
					r.Body = gzBody{gzReader, r.Body}
				}
			}
			h(w, r)
		})
	})
}

func GzipEncodeDecorator() Decorator {
	return Decorator(func(h http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				w.Header().Set("Content-Encoding", "gzip")

				gz := gzip.NewWriter(w)
				defer gz.Close()

				w = gzRespWriter{Writer: gz, ResponseWriter: w}
			}
			h(w, r)
		})
	})
}
