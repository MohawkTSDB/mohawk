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
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/MohawkTSDB/mohawk/storage"
)

type RangeIntervalType int

type Alert struct {
	ID              string            `mapstructure:"id"`
	Metric          string            `mapstructure:"metric"`
	Tenant          string            `mapstructure:"tenant"`
	From            float64           `mapstructure:"from"`
	To              float64           `mapstructure:"to"`
	Type            RangeIntervalType `mapstructure:"type"`
	State           bool
	TrigerValue     float64
	TrigerTimestamp int64
}

type AlertRules struct {
	Storage        storage.Storage
	ServerURL      string
	Verbose        bool
	Alerts         []*Alert
	AlertsInterval int
	Heartbeat      int64
}

const (
	OUTSIDE RangeIntervalType = iota
	HIGHER_THAN
	LOWER_THAN
)

// Alert

// updateAlertState update the alert status
func (alert *Alert) updateAlertState(value float64) {
	// valid metric values are:
	//    from < value >= to
	//    values outside this range will triger an alert
	switch alert.Type {
	case OUTSIDE:
		alert.State = value <= alert.From || value > alert.To
	case HIGHER_THAN:
		alert.State = value > alert.To
	case LOWER_THAN:
		alert.State = value <= alert.From
	}
}

// AlertRules

// Init fill in missing configuration data, and start the alert checking loop
func (a *AlertRules) Init() {
	// if user omit the tenant field in the alerts config, fallback to default
	// tenant
	for _, alert := range a.Alerts {
		// fall back to _ops
		if alert.Tenant == "" {
			alert.Tenant = "_ops"
		}
	}

	// check for alerts periodically.
	log.Printf("Start alerts, alert buffer url: %+v", a.ServerURL)
	log.Printf("Start alerts, alert interval: %+vs", a.AlertsInterval)
	go a.run()
}

// check for alert status chenge periodically
func (a *AlertRules) run() {
	c := time.Tick(time.Second * time.Duration(a.AlertsInterval))

	for range c {
		a.checkAlerts()
	}
}

// loop on all alerts and check for status change
func (a *AlertRules) checkAlerts() {
	var end int64
	var start int64
	var tenant string
	var metric string
	var oldState bool

	// check only for current data
	end = int64(time.Now().UTC().Unix() * 1000)
	start = end - 60*60*1000

	for _, alert := range a.Alerts {
		// check out values for the alert metric
		tenant = alert.Tenant
		metric = alert.Metric
		rawData := a.Storage.GetRawData(tenant, metric, end, start, 1, "DESC")

		// if we have new data check for alert status change
		// [ if no new data found, leave alert status un changed ]
		if len(rawData) > 0 {
			oldState = alert.State

			// update alert state and report to user if changed.
			alert.updateAlertState(rawData[0].Value)
			if alert.State != oldState {
				// set triger values
				alert.TrigerValue = rawData[0].Value
				alert.TrigerTimestamp = rawData[0].Timestamp

				if b, err := json.Marshal(alert); err == nil {
					s := string(b)
					log.Println(s)
					a.post(s)
				}
			}
		}
	}

	// update check compleat heart beat timestamp
	a.Heartbeat = end
}

func (a *AlertRules) post(s string) {
	client := http.Client{}
	req, err := http.NewRequest("POST", a.ServerURL, bytes.NewBufferString(s))
	if err == nil {
		client.Do(req)
	}
}
