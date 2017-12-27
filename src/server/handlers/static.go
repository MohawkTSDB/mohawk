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

// Package handler http server handler functions
package handler

import (
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

// Static handler to serve static file
type Static struct {
	MediaPath string
	next      http.Handler
}

// SetNext set next http serve func
func (s *Static) SetNext(h http.Handler) {
	s.next = h
}

// ServeHTTP http serve func
func (s Static) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := s.MediaPath + r.URL.EscapedPath()

	// Add index.html to path if it ends with /
	if path[len(path)-1:] == "/" {
		path = path + "index.html"
	}

	// Add /index.html to path if a directory
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		path = path + "/index.html"
	}

	// If file exist serve it
	if file, err := ioutil.ReadFile(path); err == nil {
		ext := filepath.Ext(path)

		w.Header().Set("Content-Type", mime.TypeByExtension(ext))
		w.WriteHeader(200)
		w.Write(file)

		return
	}

	// If we can not show file, continue
	s.next.ServeHTTP(w, r)
}
