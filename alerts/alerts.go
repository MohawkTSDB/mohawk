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
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/MohawkTSDB/mohawk/storage"
)

type RangeIntervalType int

type Alert struct {
	ID              string   `mapstructure:"id"`
	Metric          string   `mapstructure:"metric"`
	Tenant          string   `mapstructure:"tenant"`
	High            *float64 `mapstructure:"alert-if-higher-then"`
	Low             *float64 `mapstructure:"alert-if-lower-then"`
	Type            RangeIntervalType
	State           bool
	TrigerValue     float64
	TrigerTimestamp int64
}

type AlertRules struct {
	Storage        storage.Storage
	ServerURL      string
	ServerMethod   string
	ServerInsecure bool
	Verbose        bool
	Alerts         []*Alert
	AlertsInterval int
	Heartbeat      int64
}

const (
	NONE RangeIntervalType = iota
	OUTSIDE
	HIGHER_THAN
	LOWER_THAN
)

// Alert

// updateAlertState update the alert status
func (alert *Alert) updateAlertState(value float64) {
	if alert.Type == NONE {
		return
	}

	// valid metric values are:
	//    Low <= value > High
	//    values outside this range will triger an alert
	switch alert.Type {
	case OUTSIDE:
		alert.State = value < *alert.Low || value >= *alert.High
	case HIGHER_THAN:
		alert.State = value >= *alert.High
	case LOWER_THAN:
		alert.State = value < *alert.Low
	}
}

// AlertRules

// Init fill in missing configuration data, and start the alert checking loop
func (a *AlertRules) Init() {
	// Init alert objects
	for _, alert := range a.Alerts {
		// alert type:
		//   if we have only Low  -> alert type is lower then
		//   if we have only High -> alert type is higher then
		//   o/w                  -> alert if outside
		if alert.High == nil && alert.Low != nil {
			alert.Type = LOWER_THAN
		} else if alert.High != nil && alert.Low == nil {
			alert.Type = HIGHER_THAN
		} else if alert.High != nil && alert.Low != nil {
			alert.Type = OUTSIDE
		} else {
			alert.Type = NONE
		}

		// fall back to _ops if no tenant given
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
					a.post(s)

					// log alert status change
					if a.Verbose {
						log.Println(s)
					}
				}
			}
		}
	}

	// update check compleat heart beat timestamp
	a.Heartbeat = end
}

func (a *AlertRules) post(s string) {
	var client http.Client

	if a.ServerInsecure {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = http.Client{Transport: tr}
	} else {
		client = http.Client{}
	}

	req, e := http.NewRequest(a.ServerMethod, a.ServerURL, bytes.NewBufferString(s))
	if e == nil {
		req.Header.Set("Content-Type", "application/json")
		_, err := client.Do(req)

		// log post errors
		if a.Verbose && err != nil {
			log.Println(err)
		}
	}
}

func (alerts *AlertRules) FilterAlerts(tenant string, id string, state string) []Alert {
	res := make([]Alert, 0)
	s := state == "T"

	for _, a := range alerts.Alerts {
		if a.Tenant == tenant && (id == "" || a.ID == id) && (state == "" || a.State == s) {
			res = append(res, *a)
		}
	}

	return res
}
