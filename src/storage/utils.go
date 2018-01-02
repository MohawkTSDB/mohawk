// Copyright 2016,2017,2018 Yaacov Zamir <kobi.zamir@gmail.com>
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

// Package storage interface for metric data storage
package storage

import (
	"log"
	"strconv"
)

// FilterItems filters a list using a filter function
func FilterItems(vs []Item, f func(Item) bool) []Item {
	vsf := make([]Item, 0)

	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

// ParseSec parse a time string into seconds,
// posible postfix - s, mn, h, d
// e.g. "2h" => 2 * 60 * 60
func ParseSec(t string) int64 {
	var err error
	var i int

	if len(t) < 2 {
		log.Fatal("Can't parse time ", t)
	}

	// check for ms and mn
	switch t[len(t)-2:] {
	case "mn":
		if i, err = strconv.Atoi(t[:len(t)-2]); err == nil {
			return int64(i) * 60
		}
	}

	// check for s, h and d
	switch t[len(t)-1:] {
	case "s":
		if i, err = strconv.Atoi(t[:len(t)-1]); err == nil {
			return int64(i)
		}
	case "h":
		if i, err = strconv.Atoi(t[:len(t)-1]); err == nil {
			return int64(i) * 60 * 60
		}
	case "d":
		if i, err = strconv.Atoi(t[:len(t)-1]); err == nil {
			return int64(i) * 60 * 60 * 24
		}
	}

	// if here must be an error
	log.Fatal("Can't parse time ", t)

	return 0
}
