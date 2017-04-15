// Copyright 2016 Red Hat, Inc. and/or its affiliates
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
package random

import (
	"fmt"
	"math/rand"

	"github.com/yaacov/mohawk/backend"
)

type Backend struct {
	Items []backend.Item
}

func (r Backend) Name() string {
	return "Backend-Random"
}

func (r *Backend) Open() {
	r.Items = make([]backend.Item, 0)

	seeds := []map[string]string{
		map[string]string{"type": "node", "group_id": "cpu/usage_rate", "units": "cpu", "issue": "42"},
		map[string]string{"type": "node", "group_id": "memory/usage_rate", "units": "byte"},
		map[string]string{"type": "node", "group_id": "cpu/usage_rate", "units": "cpu", "issue": "42"},
		map[string]string{"type": "node", "group_id": "memory/usage_rate", "units": "byte"},
		map[string]string{"type": "node", "group_id": "cpu/limit", "units": "cpu", "issue": "442"},
		map[string]string{"type": "node", "group_id": "memory/limit", "units": "byte", "issue": "442"},
		map[string]string{"type": "node", "group_id": "filesystem/usage_rate", "units": "byte"},
	}

	for i := 0; i < 120; i++ {
		seed := seeds[rand.Intn(len(seeds))]
		tags := map[string]string{
			"type":     seed["type"],
			"group_id": seed["group_id"],
			"units":    seed["units"],
			"issue":    seed["issue"],
			"hostname": fmt.Sprintf("example.%03d.com", i/4),
		}
		id := fmt.Sprintf("machine/%s/%s", tags["hostname"], tags["group_id"])

		validID := true
		for _, o := range r.Items {
			if id == o.Id {
				validID = false
			}
		}

		if validID {
			r.Items = append(r.Items, backend.Item{
				Id:   id,
				Type: "gauge",
				Tags: tags,
			})
		}
	}
}

func (r Backend) GetItemList(tags map[string]string) []backend.Item {
	res := r.Items

	if len(tags) > 0 {
		for key, value := range tags {
			res = backend.FilterItems(res, func(i backend.Item) bool { return i.Tags[key] == value })
		}
	}

	return res
}

func (r Backend) GetRawData(id string, end int64, start int64, limit int64, order string) []backend.DataItem {
	var i int64
	var res []backend.DataItem

	delta := int64(5 * 60 * 1000)
	base := 10 + rand.Intn(250)
	variant := 20 + rand.Intn(50)

	for i = 0; i < limit && (end-i*delta) > start; i++ {
		res = append(res, backend.DataItem{
			Timestamp: end - i*delta,
			Value:     float64(base + rand.Intn(variant)),
		})
	}

	if order == "ASC" {
		for i, j := 0, len(res)-1; i < j; i, j = i+1, j-1 {
			res[i], res[j] = res[j], res[i]
		}
	}

	return res
}

func (r Backend) GetStatData(id string, end int64, start int64, limit int64, order string, bucketDuration int64) []backend.StatItem {
	var i int64
	var res []backend.StatItem

	delta := bucketDuration * 1000
	base := 10 + rand.Intn(250)
	variant := 20 + rand.Intn(50)

	for i = 0; i < limit && (end-(i-1)*delta) > start; i++ {
		value := float64(base + rand.Intn(variant))
		res = append(res, backend.StatItem{
			Start:   end - (i-1)*delta,
			End:     end - i*delta,
			Empty:   false,
			Samples: 1,
			Min:     value,
			Max:     value,
			Avg:     value,
			Median:  value,
			Sum:     value,
		})
	}

	if order == "ASC" {
		for i, j := 0, len(res)-1; i < j; i, j = i+1, j-1 {
			res[i], res[j] = res[j], res[i]
		}
	}

	return res
}

func (r Backend) PostRawData(id string, t int64, v float64) bool {
	return false
}

func (r Backend) PutTags(id string, tags map[string]string) bool {
	return false
}
