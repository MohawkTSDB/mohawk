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

// Package backend
package example

import (
	"fmt"
	"math/rand"
	"net/url"

	"github.com/yaacov/mohawk/backend"
)

type Backend struct {
}

// Backend functions
// Required by backend interface

func (r Backend) Name() string {
	return "Backend-Example"
}

func (r *Backend) Open(options url.Values) {
	// open db connection
}

func (r Backend) GetTenants() []backend.Tenant {
	res := make([]backend.Tenant, 0)

	// return a list of tenants
	res = append(res, backend.Tenant{Id: "Example tenant"})

	return res
}

func (r Backend) GetItemList(tenant string, tags map[string]string) []backend.Item {
	res := make([]backend.Item, 0)
	maxSize := 42

	for i := 0; i < maxSize; i++ {
		res = append(res, backend.Item{
			Id:   fmt.Sprintf("container/%08d/example/gouge", i),
			Type: "gauge",
			Tags: map[string]string{"name": "example/gouge", "units": "byte"},
		})
	}

	// filter using tags
	// 	if we have a list of _all_ items, we need to filter them by tags
	// 	if the list is already filtered, we do not need to re-filter it
	/*
		if len(tags) > 0 {
			for key, value := range tags {
				res = backend.FilterItems(res, func(i backend.Item) bool {
					r, _ := regexp.Compile("^" + value + "$")
					return r.MatchString(i.Tags[key])
				})
			}
		}
	*/

	return res
}

func (r Backend) GetRawData(tenant string, id string, end int64, start int64, limit int64, order string) []backend.DataItem {
	res := make([]backend.DataItem, 0)
	var sampleDuration int64
	var l int64
	var i int64

	sampleDuration = 5
	l = (end - start) / 1000 / sampleDuration
	if limit < l {
		l = limit
	}

	for i = 0; i < l; i++ {
		res = append(res, backend.DataItem{
			Timestamp: end - sampleDuration*1000*i,
			Value:     124 + float64(rand.Intn(42)),
		})
	}

	return res
}

func (r Backend) GetStatData(tenant string, id string, end int64, start int64, limit int64, order string, bucketDuration int64) []backend.StatItem {
	res := make([]backend.StatItem, 0)
	var l int64
	var i int64

	l = (end - start) / 1000 / bucketDuration
	if limit < l {
		l = limit
	}

	for i = 0; i < l; i++ {
		res = append(res, backend.StatItem{
			Start:   end - bucketDuration*1000*(i+1),
			End:     end - bucketDuration*1000*i,
			Empty:   false,
			Samples: 1,
			Min:     0,
			Max:     0,
			First:   0,
			Last:    0,
			Avg:     124 + float64(rand.Intn(42)),
			Median:  0,
			Std:     0,
			Sum:     0,
		})
	}

	return res
}

// unimplemented requests should fail silently

func (r Backend) PostRawData(tenant string, id string, t int64, v float64) bool {
	return true
}

func (r Backend) PutTags(tenant string, id string, tags map[string]string) bool {
	return true
}

func (r Backend) DeleteData(tenant string, id string, end int64, start int64) bool {
	return true
}

func (r Backend) DeleteTags(tenant string, id string, tags []string) bool {
	return true
}

// Helper functions
// Not required by backend interface
