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
package memory

import (
	"net/url"
	"regexp"
	"time"

	"github.com/yaacov/mohawk/backend"
)

type TimeValuePair struct {
	timeStamp int64
	value     float64
}

type TimeSeries struct {
	tags map[string]string
	data []TimeValuePair
}

type Tenant struct {
	ts map[string]*TimeSeries
}

type Backend struct {
	firstPosTimestamp  int64
	timeGranularitySec int64

	tenant map[string]*Tenant
}

// Backend functions
// Required by backend interface

func (r Backend) Name() string {
	return "Backend-Memory"
}

func (r *Backend) Open(options url.Values) {
	// set time granularity to 5 sec
	r.timeGranularitySec = 5
	r.firstPosTimestamp = int64(time.Now().UTC().Unix() / r.timeGranularitySec)

	// open db connection
	r.tenant = make(map[string]*Tenant, 0)
}

func (r Backend) GetTenants() []backend.Tenant {
	res := make([]backend.Tenant, 0)

	// return a list of tenants
	for key, _ := range r.tenant {
		res = append(res, backend.Tenant{Id: key})
	}

	return res
}

func (r Backend) GetItemList(tenant string, tags map[string]string) []backend.Item {
	res := make([]backend.Item, 0)
	t, ok := r.tenant[tenant]

	// check tenant
	if !ok {
		return res
	}

	for key, ts := range t.ts {
		res = append(res, backend.Item{
			Id:   key,
			Type: "gauge",
			Tags: ts.tags,
		})
	}

	// filter using tags
	// 	if we have a list of _all_ items, we need to filter them by tags
	// 	if the list is already filtered, we do not need to re-filter it
	if len(tags) > 0 {
		for key, value := range tags {
			res = backend.FilterItems(res, func(i backend.Item) bool {
				r, _ := regexp.Compile("^" + value + "$")
				return r.MatchString(i.Tags[key])
			})
		}
	}

	return res
}

func (r *Backend) GetRawData(tenant string, id string, end int64, start int64, limit int64, order string) []backend.DataItem {
	res := make([]backend.DataItem, 0)

	// check if tenant and id exists, create them if necessary
	r.checkID(tenant, id)

	for _, v := range r.tenant[tenant].ts[id].data {
		res = append(res, backend.DataItem{
			Timestamp: v.timeStamp * 1000,
			Value:     v.value,
		})
	}

	return res
}

func (r Backend) GetStatData(tenant string, id string, end int64, start int64, limit int64, order string, bucketDuration int64) []backend.StatItem {
	var res []backend.StatItem

	res = append(res, backend.StatItem{
		Start:   start,
		End:     end,
		Empty:   true,
		Samples: 0,
		Min:     0,
		Max:     0,
		Avg:     0,
		Median:  0,
		Sum:     0,
	})

	return res
}

func (r *Backend) PostRawData(tenant string, id string, t int64, v float64) bool {
	// check if tenant and id exists, create them if necessary
	r.checkID(tenant, id)

	// update time value pair to the time serias
	r.tenant[tenant].ts[id].data = append(r.tenant[tenant].ts[id].data, TimeValuePair{timeStamp: t, value: v})

	return true
}

func (r *Backend) PutTags(tenant string, id string, tags map[string]string) bool {
	// check if tenant and id exists, create them if necessary
	r.checkID(tenant, id)

	// update time serias tags
	if len(tags) > 0 {
		for key, value := range tags {
			r.tenant[tenant].ts[id].tags[key] = value
		}
	}

	return true
}

func (r *Backend) DeleteData(tenant string, id string, end int64, start int64) bool {
	return true
}

func (r *Backend) DeleteTags(tenant string, id string, tags []string) bool {
	return true
}

// Helper functions
// Not required by backend interface

func (r *Backend) checkID(tenant string, id string) {
	var ok bool

	// check for tenant
	if _, ok = r.tenant[tenant]; !ok {
		r.tenant[tenant] = &Tenant{ts: make(map[string]*TimeSeries)}
	}

	// check for TimeSeries
	if _, ok = r.tenant[tenant].ts[id]; !ok {
		r.tenant[tenant].ts[id] = &TimeSeries{
			tags: make(map[string]string),
			data: make([]TimeValuePair, 10),
		}
	}
}
