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
package backend

import (
	"fmt"
	"math/rand"
)

type Random struct {
	Items []Item
}

func (r Random) Name() string {
	return "Backend-Random"
}

func (r *Random) Open() {
	r.Items = make([]Item, 0)

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
		r.Items = append(r.Items, Item{
			Id:   fmt.Sprintf("hello_kitty_%03d", i),
			Type: "gauge",
			Tags: tags,
		})
	}
}

func (r Random) GetItemList(tags map[string]string) []Item {
	res := r.Items

	if len(tags) > 0 {
		for key, value := range tags {
			res = FilterItems(res, func(i Item) bool { return i.Tags[key] == value })
		}
	}

	return res
}

func (r Random) GetRawData(id string, end int64, start int64, limit int64, order string) []DataItem {
	res := make([]DataItem, 0)

	delta := int64(5 * 60 * 1000)

	for i := limit; i > 0; i-- {
		res = append(res, DataItem{
			Timestamp: end - i*delta,
			Value:     float64(50 + rand.Intn(50)),
		})
	}

	if order == "DESC" {
		for i, j := 0, len(res)-1; i < j; i, j = i+1, j-1 {
			res[i], res[j] = res[j], res[i]
		}
	}

	return res
}

func (r Random) GetStatData(id string, end int64, start int64, limit int64, order string, bucketDuration int64) []StatItem {
	res := make([]StatItem, 0)

	delta := int64(5 * 60 * 1000)

	for i := limit; i > 0; i-- {
		value := float64(50 + rand.Intn(50))
		res = append(res, StatItem{
			Start:   end - i*delta,
			End:     end - (i-1)*delta,
			Empty:   false,
			Samples: 1,
			Min:     value,
			Max:     value,
			Avg:     value,
			Median:  value,
			Sum:     value,
		})
	}

	if order == "DESC" {
		for i, j := 0, len(res)-1; i < j; i, j = i+1, j-1 {
			res[i], res[j] = res[j], res[i]
		}
	}

	return res
}

func (r Random) PostRawData(id string, t int64, v float64) bool {
	return false
}

func (r Random) PutTags(id string, tags map[string]string) bool {
	return false
}
