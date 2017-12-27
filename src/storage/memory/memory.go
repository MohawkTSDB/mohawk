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

// Package storage
package memory

import (
	"log"
	"net/url"
	"regexp"
	"time"

	"github.com/MohawkTSDB/mohawk/src/storage"
)

type TimeValuePair struct {
	timeStamp int64
	value     float64
}

type TimeSeries struct {
	tags      map[string]string
	data      []TimeValuePair
	lastValue TimeValuePair
}

type Tenant struct {
	ts map[string]*TimeSeries
}

type Storage struct {
	timeGranularitySec int64
	timeRetentionSec   int64
	timeLastSec        int64

	tenant map[string]*Tenant
}

// Storage functions
// Required by storage interface

func (r Storage) Name() string {
	return "Storage-Memory"
}

func (r *Storage) Open(options url.Values) {
	// set last entry time
	r.timeLastSec = 0
	// set time granularity to 30 sec
	r.timeGranularitySec = 30
	// set time retention to 7 days
	r.timeRetentionSec = 7 * 24 * 60 * 60

	// open db connection
	r.tenant = make(map[string]*Tenant, 0)

	// start a maintenance worker that will clean the db periodically
	go r.maintenance()
}

func (r Storage) GetTenants() []storage.Tenant {
	res := make([]storage.Tenant, 0, len(r.tenant))

	// return a list of tenants
	for key := range r.tenant {
		res = append(res, storage.Tenant{ID: key})
	}

	return res
}

func (r Storage) GetItemList(tenant string, tags map[string]string) []storage.Item {
	res := make([]storage.Item, 0)
	t, ok := r.tenant[tenant]

	// check tenant
	if !ok {
		return res
	}

	for key, ts := range t.ts {
		if hasMatchingTag(tags, ts.tags) {
			lastValue := storage.DataItem{
				Timestamp: ts.lastValue.timeStamp,
				Value:     ts.lastValue.value,
			}

			res = append(res, storage.Item{
				ID:         key,
				Type:       "gauge",
				Tags:       ts.tags,
				LastValues: []storage.DataItem{lastValue},
			})
		}
	}

	return res
}

func (r *Storage) GetRawData(tenant string, id string, end int64, start int64, limit int64, order string) []storage.DataItem {
	res := make([]storage.DataItem, 0)

	arraySize := r.timeRetentionSec / r.timeGranularitySec
	pStart := r.getPosForTimestamp(start)
	pEnd := r.getPosForTimestamp(end)

	// make sure start and end times is in the retention time
	start, end = r.checkTimespan(start, end)

	// sanity check pEnd
	if pEnd <= pStart {
		pEnd += arraySize
	}

	// check if tenant and id exists, create them if necessary
	r.checkID(tenant, id)

	// fill data out array
	count := int64(0)
	for i := pEnd; count < limit && i >= pStart; i-- {
		d := r.tenant[tenant].ts[id].data[i%arraySize]

		// if this is a valid point
		if d.timeStamp < end && d.timeStamp >= start {
			count++
			res = append(res, storage.DataItem{
				Timestamp: d.timeStamp,
				Value:     d.value,
			})
		}
	}

	// order
	if order == "ASC" {
		for i := 0; i < len(res)/2; i++ {
			j := len(res) - i - 1
			res[i], res[j] = res[j], res[i]
		}
	}

	return res
}

func (r Storage) GetStatData(tenant string, id string, end int64, start int64, limit int64, order string, bucketDuration int64) []storage.StatItem {
	res := make([]storage.StatItem, 0)

	// make sure start and end times is in the retention time
	start, end = r.checkTimespan(start, end)

	// bucketDuration can't be smaller then granularity
	if bucketDuration < r.timeGranularitySec {
		bucketDuration = r.timeGranularitySec
	}
	bucketDuration = r.timeGranularitySec * (bucketDuration / r.timeGranularitySec)

	// start and tend must be integer multiplections of bucketDuration
	bucketDurationMilli := bucketDuration * 1000
	start = bucketDurationMilli * (start / bucketDurationMilli)
	end = bucketDurationMilli * (1 + end/bucketDurationMilli)

	arraySize := r.timeRetentionSec / r.timeGranularitySec
	pStep := bucketDuration / r.timeGranularitySec
	pStart := r.getPosForTimestamp(start)
	pEnd := r.getPosForTimestamp(end)

	// sanity check pEnd
	if pEnd <= pStart {
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
	for b := pEnd; count < limit && b > pStart && startTimestamp >= stepMillisec; b -= pStep {
		samples := int64(0)
		sum := float64(0)
		first := float64(0)
		last := float64(0)
		min := float64(0)
		max := float64(0)

		// loop on all points in bucket
		for i := (b - pStep); i < b; i++ {
			d := r.tenant[tenant].ts[id].data[i%arraySize]
			if d.timeStamp <= end && d.timeStamp > start {
				samples++

				// calculate bucket stat values
				if samples == 1 {
					// first sample
					first = d.value
					min = first
					max = first
				} else {
					// all samples except first sample
					if min > d.value {
						min = d.value
					}
					if max < d.value {
						max = d.value
					}
				}

				last = d.value
				sum += d.value
			}
		}

		// all points are valid
		startTimestamp -= stepMillisec
		count++

		// all points are valid
		if samples > 0 {
			res = append(res, storage.StatItem{
				Start:   startTimestamp,
				End:     startTimestamp + stepMillisec,
				Empty:   false,
				Samples: samples,
				First:   first,
				Last:    last,
				Min:     min,
				Max:     max,
				Avg:     sum / float64(samples),
				Sum:     sum,
			})
		} else {
			count++
			res = append(res, storage.StatItem{
				Start: startTimestamp,
				End:   startTimestamp + stepMillisec,
				Empty: true,
			})
		}
	}

	// order
	if order == "ASC" {
		for i := 0; i < len(res)/2; i++ {
			j := len(res) - i - 1
			res[i], res[j] = res[j], res[i]
		}
	}

	return res
}

func (r *Storage) PostRawData(tenant string, id string, t int64, v float64) bool {
	// check if tenant and id exists, create them if necessary
	r.checkID(tenant, id)

	// update time value pair to the time serias
	p := r.getPosForTimestamp(t)
	r.tenant[tenant].ts[id].data[p] = TimeValuePair{timeStamp: t, value: v}

	// update last value
	if r.tenant[tenant].ts[id].lastValue.timeStamp < t {
		r.tenant[tenant].ts[id].lastValue.timeStamp = t
		r.tenant[tenant].ts[id].lastValue.value = v
	}

	// update last
	tSec := t / 1000
	if tSec > r.timeLastSec {
		r.timeLastSec = tSec
	}

	return true
}

func (r *Storage) PutTags(tenant string, id string, tags map[string]string) bool {
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

func (r *Storage) DeleteData(tenant string, id string, end int64, start int64) bool {
	return true
}

func (r *Storage) DeleteTags(tenant string, id string, tags []string) bool {
	return true
}

// Helper functions
// Not required by storage interface

func (r *Storage) checkTimespan(start int64, end int64) (int64, int64) {
	memFirstTime := (r.timeLastSec - r.timeRetentionSec) * 1000
	memLastTime := r.timeLastSec * 1000

	// make sure start and end times is in the retention time
	if start < memFirstTime {
		start = memFirstTime
	}
	if end > memLastTime {
		end = memLastTime + 1
	}

	return start, end
}

func (r *Storage) getPosForTimestamp(timestamp int64) int64 {
	arraySize := r.timeRetentionSec / r.timeGranularitySec
	arrayPos := timestamp / 1000 / r.timeGranularitySec

	return arrayPos % arraySize
}

func (r *Storage) checkID(tenant string, id string) {
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

func hasMatchingTag(tags map[string]string, itemTags map[string]string) bool {
	out := true

	// if no tags, all items match
	if len(tags) == 0 {
		return true
	}

	// if item has no tags, item is invalid
	if len(itemTags) == 0 {
		return false
	}

	// loop on all the tags, we need _all_ query tags to match tags on item
	for key, value := range tags {
		r, _ := regexp.Compile("^" + value + "$")
		out = out && r.MatchString(itemTags[key])
	}

	return out
}

func (r *Storage) maintenance() {
	// clean data every 120 minutes
	c := time.Tick(120 * time.Minute)

	// once a tick clean data
	for range c {
		log.Printf("maintenance: start\n")
		r.cleanData()
	}
}

func (r *Storage) cleanData() {
	var lastTimeStampSec int64
	validTimeStamp := time.Now().Unix() - r.timeRetentionSec

	// loop on all tenants
	for _, t := range r.tenant {
		// loop on all time series in this tenant
		for key, ts := range t.ts {
			lastTimeStampSec = ts.lastValue.timeStamp / 1000

			// if last value is more then time span old, remove data
			if lastTimeStampSec <= validTimeStamp {
				log.Printf("maintenance: delete item %s\n", key)
				delete(t.ts, key)
			}
		}

		// TODO: delete tenant if no time seriess
	}
}
