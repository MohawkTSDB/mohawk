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

// Package storage interface for metric data storage
package example

import (
	"fmt"
	"math/rand"
	"net/url"

	"github.com/MohawkTSDB/mohawk/src/storage"
)

type Storage struct {
}

// Storage functions
// Required by storage interface

func (r Storage) Name() string {
	return "Storage-Example"
}

func (r *Storage) Open(options url.Values) {
	// open db connection
}

func (r Storage) GetTenants() []storage.Tenant {
	res := make([]storage.Tenant, 0)

	// return a list of tenants
	res = append(res, storage.Tenant{ID: "Example tenant"})

	return res
}

func (r Storage) GetItemList(tenant string, tags map[string]string) []storage.Item {
	res := make([]storage.Item, 0)
	maxSize := 42

	for i := 0; i < maxSize; i++ {
		res = append(res, storage.Item{
			ID:   fmt.Sprintf("container/%08d/example/gouge", i),
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
				res = storage.FilterItems(res, func(i storage.Item) bool {
					r, _ := regexp.Compile("^" + value + "$")
					return r.MatchString(i.Tags[key])
				})
			}
		}
	*/

	return res
}

func (r Storage) GetRawData(tenant string, id string, end int64, start int64, limit int64, order string) []storage.DataItem {
	res := make([]storage.DataItem, 0)
	var sampleDuration int64
	var l int64
	var i int64

	sampleDuration = 5
	l = (end - start) / 1000 / sampleDuration
	if limit < l {
		l = limit
	}

	for i = 0; i < l; i++ {
		res = append(res, storage.DataItem{
			Timestamp: end - sampleDuration*1000*i,
			Value:     124 + float64(rand.Intn(42)),
		})
	}

	return res
}

func (r Storage) GetStatData(tenant string, id string, end int64, start int64, limit int64, order string, bucketDuration int64) []storage.StatItem {
	res := make([]storage.StatItem, 0)
	var l int64
	var i int64

	l = (end - start) / 1000 / bucketDuration
	if limit < l {
		l = limit
	}

	for i = 0; i < l; i++ {
		res = append(res, storage.StatItem{
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

func (r Storage) PostRawData(tenant string, id string, t int64, v float64) bool {
	return true
}

func (r Storage) PutTags(tenant string, id string, tags map[string]string) bool {
	return true
}

func (r Storage) DeleteData(tenant string, id string, end int64, start int64) bool {
	return true
}

func (r Storage) DeleteTags(tenant string, id string, tags []string) bool {
	return true
}

// Helper functions
// Not required by storage interface
