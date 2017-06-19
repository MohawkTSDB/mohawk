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
	"fmt"
	"net/url"
	"regexp"

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
	timeGranularitySec int64
	timeRetentionSec   int64
	timeLastSec        int64

	tenant map[string]*Tenant
}

// Backend functions
// Required by backend interface

func (r Backend) Name() string {
	return "Backend-Memory"
}

func (r *Backend) Open(options url.Values) {
	// set last entry time
	r.timeLastSec = 0
	// set time granularity to 30 sec
	r.timeGranularitySec = 30
	// set time retention to 7 days
	r.timeRetentionSec = 7 * 24 * 60 * 60

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
	memFirstTime := (r.timeLastSec - r.timeRetentionSec) * 1000
	memLastTime := r.timeLastSec * 1000

	arraySize := r.timeRetentionSec / r.timeGranularitySec
	pStart := r.getPosForTimestamp(start)
	pEnd := r.getPosForTimestamp(end)

	// make sure start and end times is in the retention time
	if start < memFirstTime {
		start = memFirstTime
	}
	if end > memLastTime {
		end = memLastTime + 1
	}

	// sanity check pEnd
	if pEnd < pStart {
		pEnd += arraySize
	}

	// check if tenant and id exists, create them if necessary
	r.checkID(tenant, id)

	// fill data out array
	count := int64(0)
	for i := pStart; count < limit && i <= pEnd; i++ {
		d := r.tenant[tenant].ts[id].data[i%arraySize]

		// if this is a valid point
		if d.timeStamp < end && d.timeStamp >= start {
			count++
			res = append(res, backend.DataItem{
				Timestamp: d.timeStamp,
				Value:     d.value,
			})
		}
	}

	return res
}

func (r Backend) GetStatData(tenant string, id string, end int64, start int64, limit int64, order string, bucketDuration int64) []backend.StatItem {
	res := make([]backend.StatItem, 0)
	memFirstTime := (r.timeLastSec - r.timeRetentionSec) * 1000
	memLastTime := r.timeLastSec * 1000

	// make sure start and end times is in the retention time
	if start < memFirstTime {
		start = memFirstTime
	}
	if end > memLastTime {
		end = memLastTime + 1
	}

	// make sure start, end and backetDuration is a multiple of granularity
	bucketDuration = r.timeGranularitySec * (1 + bucketDuration/r.timeGranularitySec)
	start = r.timeGranularitySec * (1 + start/1000/r.timeGranularitySec) * 1000
	end = r.timeGranularitySec * (1 + end/1000/r.timeGranularitySec) * 1000
	fmt.Printf("stat: %d %d %d", bucketDuration, start, end)
	arraySize := r.timeRetentionSec / r.timeGranularitySec
	pStep := bucketDuration / r.timeGranularitySec
	pStart := r.getPosForTimestamp(start)
	pEnd := r.getPosForTimestamp(end)

	// sanity check pEnd
	if pEnd < pStart {
		pEnd += arraySize
	}

	// sanity check step
	if pStep < 1 {
		pStep = 1
	}
	if pStep > (pEnd - pStart) {
		pStep = pEnd - pStart
	}

	startTimestamp := end
	stepMillisec := pStep * r.timeGranularitySec * 1000

	// check if tenant and id exists, create them if necessary
	r.checkID(tenant, id)

	// fill data out array
	count := int64(0)
	for b := pEnd; count < limit && b > pStart && startTimestamp > stepMillisec; b -= pStep {
		samples := int64(0)
		sum := float64(0)
		last := float64(0)

		// loop on all points in bucket
		for i := (b - pStep); i < b; i++ {
			d := r.tenant[tenant].ts[id].data[i%arraySize]
			if d.timeStamp <= end && d.timeStamp > start {
				samples++
				last = d.value
				sum += d.value
			}
		}

		// all points are valid
		startTimestamp -= stepMillisec
		count++

		// all points are valid
		if samples > 0 {
			res = append(res, backend.StatItem{
				Start:   startTimestamp,
				End:     startTimestamp + stepMillisec,
				Empty:   false,
				Samples: samples,
				Min:     0,
				Max:     last,
				Avg:     sum / float64(samples),
				Median:  0,
				Sum:     sum,
			})
		} else {
			count++
			res = append(res, backend.StatItem{
				Start:   startTimestamp,
				End:     startTimestamp + stepMillisec,
				Empty:   true,
				Samples: 0,
				Min:     0,
				Max:     0,
				Avg:     0,
				Median:  0,
				Sum:     0,
			})
		}
	}

	return res
}

func (r *Backend) PostRawData(tenant string, id string, t int64, v float64) bool {
	// check if tenant and id exists, create them if necessary
	r.checkID(tenant, id)

	// update time value pair to the time serias
	p := r.getPosForTimestamp(t)
	r.tenant[tenant].ts[id].data[p] = TimeValuePair{timeStamp: t, value: v}
	fmt.Printf("post: %d %d %f %d", p, t, v, t/1000)
	// update last
	r.timeLastSec = t / 1000

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

func (r *Backend) getPosForTimestamp(timestamp int64) int64 {
	arraySize := r.timeRetentionSec / r.timeGranularitySec
	arrayPos := timestamp / 1000 / r.timeGranularitySec

	return arrayPos % arraySize
}

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
			data: make([]TimeValuePair, r.timeRetentionSec/r.timeGranularitySec),
		}
	}
}
