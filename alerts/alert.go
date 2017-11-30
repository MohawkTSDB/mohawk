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

// Package alerts for alert rules
package alerts

import (
	"time"
	"fmt"
	"encoding/json"

	"github.com/MohawkTSDB/mohawk/storage"
)

const (
	BETWEEN RangeIntervalType = iota
	HIGHER_THAN
	LOWER_THAN
)

type RangeIntervalType int

type Alert struct {
	ID      string            `mapstructure:"id"`
	Metric  string            `mapstructure:"metric"`
	Tenant  string            `mapstructure:"tenant"`
	State   bool              `mapstructure:"state"`
	From    float64           `mapstructure:"from"`
	To      float64           `mapstructure:"to"`
	Type    RangeIntervalType `mapstructure:"type"`
}

type Alerts struct{
	Backend storage.Backend
	Verbose bool
	Alerts  []*Alert
}

// Init fill in missing configuration data, and start the alert checking loop
func (a *Alerts) Init() {
	// if user omit the tenant field in the alerts config, fallback to default
	// tenant
	for _, alert := range a.Alerts {
		// fall back to _ops
		if alert.Tenant == "" {
			alert.Tenant = "_ops"
		}
	}

	// check for alerts periodically.
	go a.run()
}

// check for alert status chenge periodically
func (a *Alerts) run() {
	c := time.Tick(time.Second * 10)

	for range c {
		fmt.Printf("alert check: start\n")
		a.checkAlerts()
	}
}

// updateAlertState update the alert status
func (alert *Alert) updateAlertState(value float64) {
	// valid metric values are:
	//    from < value >= to
	//    values outside this range will triger an alert
	switch alert.Type {
	case BETWEEN:
		alert.State = value <= alert.From || value > alert.To
	case HIGHER_THAN:
		alert.State = value > alert.To
	case LOWER_THAN:
		alert.State = value <= alert.From
	}
}

// loop on all alerts and check for status change
func (a *Alerts) checkAlerts() {
	var end      int64
	var start    int64
	var tenant   string
	var metric   string
	var oldState bool

	for _, alert := range a.Alerts {
		// look for last values
		end = int64(time.Now().UTC().Unix() * 1000)
		start = end - 60*60*1000

		// check out values for the alert metric
		tenant = alert.Tenant
		metric = alert.Metric
		rawData := a.Backend.GetRawData(tenant, metric, end, start, 1, "ASC")

		// if we have new data check for alert status change
		// [ if no new data found, leave alert status un changed ]
		if len(rawData) > 0 {
			oldState = alert.State

			// update alert state and report to user if changed.
			alert.updateAlertState(rawData[0].Value)
			if alert.State != oldState {
				if b, err := json.Marshal(alert); err == nil {
					fmt.Println(string(b))
				}
			}
		}
	}
}
