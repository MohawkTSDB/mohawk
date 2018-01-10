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

// Package example interface for example metric data storage
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

// Help return a human readable storage name
func (r Storage) Name() string {
	return "Storage-Example"
}

// Help return a human readable storage help message
func (r Storage) Help() string {
	return `Example storage [example]: no options defined.
`
}

// Open storage
func (r *Storage) Open(options url.Values) {
	// open db connection
}

func (r Storage) GetTenants() ([]storage.Tenant, error) {
	res := make([]storage.Tenant, 0)

	// return a list of tenants
	res = append(res, storage.Tenant{ID: "Example tenant"})

	return res, nil
}

func (r Storage) GetItemList(tenant string, tags map[string]string) ([]storage.Item, error) {
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

	return res, nil
}

func (r Storage) GetRawData(tenant string, id string, end int64, start int64, limit int64, order string) ([]storage.DataItem, error) {
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

	return res, nil
}

func (r Storage) GetStatData(tenant string, id string, end int64, start int64, limit int64, order string, bucketDuration int64) ([]storage.StatItem, error) {
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

	return res, nil
}

// unimplemented requests should fail silently

// PostRawData handle posting data to db
func (r Storage) PostRawData(tenant string, id string, t int64, v float64) error {
	return nil
}

// PutTags handle posting tags to db
func (r Storage) PutTags(tenant string, id string, tags map[string]string) error {
	return nil
}

// DeleteData handle delete data fron db
func (r Storage) DeleteData(tenant string, id string, end int64, start int64) error {
	return nil
}

// DeleteTags handle delete tags from db
func (r Storage) DeleteTags(tenant string, id string, tags []string) error {
	return nil
}

// Helper functions
// Not required by storage interface
