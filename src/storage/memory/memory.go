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

// Package memory interface for memory metric data storage
package memory

import (
	"errors"
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
	arraySize          int64

	tenant map[string]*Tenant
}

// Storage functions
// Required by storage interface

// Name return a human readable storage name
func (r Storage) Name() string {
	return "Storage-Memory"
}

// Help return a human readable storage help message
func (r Storage) Help() string {
	return `Memory storage [memory]:
	granularity - (optional) samples max granularity (default "30s").
	retention   - (optional) samples max retention (default "1d").
	Examples:
		--options=retention=6h&granularity=30s`
}

// Open storage
func (r *Storage) Open(options url.Values) {
	granularity := int64(30)
	retention := int64(24 * 60 * 60)

	// check for user options
	granularityStr := options.Get("granularity")
	if granularityStr != "" {
		granularity = storage.ParseSec(granularityStr)
	}
	retentionStr := options.Get("retention")
	if retentionStr != "" {
		retention = storage.ParseSec(retentionStr)
	}

	// set last entry time
	r.timeLastSec = 0
	// set time granularity to 30 sec
	r.timeGranularitySec = granularity
	// set time retention to 7 days
	r.timeRetentionSec = retention
	// calculate array size
	r.arraySize = r.timeRetentionSec / r.timeGranularitySec

	// open db connection
	r.tenant = make(map[string]*Tenant, 0)

	// log init arguments
	log.Printf("Start memory storage:")
	log.Printf("  granularity: %ds", r.timeGranularitySec)
	log.Printf("  retention: %ds", r.timeRetentionSec)

	// start a maintenance worker that will clean the db periodically
	go r.maintenance()
}

func (r Storage) GetTenants() ([]storage.Tenant, error) {
	res := make([]storage.Tenant, 0, len(r.tenant))

	// return a list of tenants
	for key := range r.tenant {
		res = append(res, storage.Tenant{ID: key})
	}

	return res, nil
}

func (r Storage) GetItemList(tenant string, tags map[string]string) ([]storage.Item, error) {
	res := make([]storage.Item, 0)
	t, ok := r.tenant[tenant]

	// check tenant
	if !ok {
		return res, errors.New("memory: Can't set tenant")
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

	return res, nil
}

func (r *Storage) GetRawData(tenant string, id string, end int64, start int64, limit int64, order string) ([]storage.DataItem, error) {
	res := make([]storage.DataItem, 0)

	pStart := r.getPosForTimestamp(start)
	pEnd := r.getPosForTimestamp(end)

	// check if tenant and id exists, create them if necessary
	r.checkID(tenant, id)

	// fill data out array
	count := int64(0)
	ts := r.tenant[tenant].ts[id]

	for i := pStart; count < limit && i <= pEnd; i++ {
		d := ts.data[i%r.arraySize]

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
	if order == "DESC" {
		for i := 0; i < len(res)/2; i++ {
			j := len(res) - i - 1
			res[i], res[j] = res[j], res[i]
		}
	}

	return res, nil
}

func (r Storage) GetStatData(tenant string, id string, end int64, start int64, limit int64, order string, bucketDuration int64) ([]storage.StatItem, error) {
	var samples int64
	var sum float64
	var first float64
	var last float64
	var min float64
	var max float64

	res := make([]storage.StatItem, 0)
	pEnd, pStart, pStep := r.getStatTimes(end, start, bucketDuration)

	// check if tenant and id exists, create them if necessary
	r.checkID(tenant, id)

	// fill data out array
	count := int64(0)
	ts := r.tenant[tenant].ts[id]
	stepMili := r.timeGranularitySec * 1000

	for b := pStart; count < limit && b <= pEnd; b = b + pStep {
		samples = 0
		sum = 0

		// loop on all points in bucket
		for i := b; i < (b + pStep); i++ {
			d := ts.data[i%r.arraySize]

			if d.timeStamp <= end && d.timeStamp >= start {
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
				sum = sum + d.value
			}
		}

		// all points are valid
		if samples > 0 {
			count++

			res = append(res, storage.StatItem{
				Start:   start + (b-pStart)*stepMili,
				End:     start + (b-pStart+pStep)*stepMili,
				Empty:   false,
				Samples: samples,
				First:   first,
				Last:    last,
				Min:     min,
				Max:     max,
				Avg:     sum / float64(samples),
				Sum:     sum,
			})
		}
	}

	// order
	if order == "DESC" {
		for i := 0; i < len(res)/2; i++ {
			j := len(res) - i - 1
			res[i], res[j] = res[j], res[i]
		}
	}

	return res, nil
}

// PostRawData handle posting data to db
func (r *Storage) PostRawData(tenant string, id string, t int64, v float64) error {
	// check if tenant and id exists, create them if necessary
	r.checkID(tenant, id)

	// update time value pair to the time serias
	// unless slot already have valid value
	p := r.getPosForTimestamp(t)
	if r.tenant[tenant].ts[id].data[p%r.arraySize].timeStamp < (t - r.timeGranularitySec*1000) {
		r.tenant[tenant].ts[id].data[p%r.arraySize] = TimeValuePair{timeStamp: t, value: v}
	}

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

	return nil
}

// PutTags handle posting tags to db
func (r *Storage) PutTags(tenant string, id string, tags map[string]string) error {
	// check if tenant and id exists, create them if necessary
	r.checkID(tenant, id)

	// update time serias tags
	if len(tags) > 0 {
		for key, value := range tags {
			r.tenant[tenant].ts[id].tags[key] = value
		}
	}

	return nil
}

// DeleteData handle delete data fron db
func (r *Storage) DeleteData(tenant string, id string, end int64, start int64) error {
	return nil
}

// DeleteTags handle delete tags fron db
func (r *Storage) DeleteTags(tenant string, id string, tags []string) error {
	return nil
}

// Helper functions
// Not required by storage interface

func (r Storage) getStatTimes(end int64, start int64, bucketDuration int64) (int64, int64, int64) {
	pStep := bucketDuration / r.timeGranularitySec
	pStart := r.getPosForTimestamp(start)
	pEnd := r.getPosForTimestamp(end)

	// sanity check step
	if pStep < 1 {
		pStep = 1
	}

	return pEnd, pStart, pStep
}

func (r *Storage) getPosForTimestamp(timestamp int64) int64 {
	return timestamp / 1000 / r.timeGranularitySec
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
