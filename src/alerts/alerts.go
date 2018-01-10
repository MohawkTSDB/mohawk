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

// Package alerts for alert rules
package alerts

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/MohawkTSDB/mohawk/src/storage"
)

// AlertType describe alert type
type AlertType int

// Alert defines one alert
type Alert struct {
	ID                string            `mapstructure:"id"`
	Annotations       map[string]string `mapstructure:"annotations"`
	Metrics           []string          `mapstructure:"metrics"`
	Tags              string            `mapstructure:"tags"`
	Tenant            string            `mapstructure:"tenant"`
	AlertIfHigherThan *float64          `mapstructure:"alert-if-higher-than"`
	AlertIfLowerThan  *float64          `mapstructure:"alert-if-lower-than"`
	Type              AlertType
	State             bool
	TrigerMetric      string
	TrigerValue       float64
	TrigerTimestamp   int64
}

// AlertRules defines prameters for the alerts run rutine
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
	undefined AlertType = iota
	outside
	higherThan
	lowerThan
)

// Alert

// alertState calculate the alert status
func (alert *Alert) alertState(value float64) bool {
	if alert.Type == undefined {
		return false
	}

	// valid metric values are:
	//    AlertIfLowerThen <= value > AlertIfHigherThen
	//    values outside this range will triger an alert
	switch alert.Type {
	case outside:
		return value < *alert.AlertIfLowerThan || value >= *alert.AlertIfHigherThan
	case higherThan:
		return value >= *alert.AlertIfHigherThan
	case lowerThan:
		return value < *alert.AlertIfLowerThan
	}

	return false
}

// AlertRules

// Init fill in missing configuration data, and start the alert checking loop
func (a *AlertRules) Init() {
	// Init alert objects
	for _, alert := range a.Alerts {
		// alert type:
		//   if we have only AlertIfLowerThen  -> alert type is lower then
		//   if we have only AlertIfHigherThen -> alert type is higher then
		//   o/w                  -> alert if outside
		if alert.AlertIfHigherThan == nil && alert.AlertIfLowerThan != nil {
			alert.Type = lowerThan
		} else if alert.AlertIfHigherThan != nil && alert.AlertIfLowerThan == nil {
			alert.Type = higherThan
		} else if alert.AlertIfHigherThan != nil && alert.AlertIfLowerThan != nil {
			alert.Type = outside
		} else {
			alert.Type = undefined
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
	var limit int64
	var tenant string
	var oldState bool
	var newState bool

	// check last 15 minutes
	end = int64(time.Now().UTC().Unix() * 1000)
	start = end - 5*60*1000
	limit = 5 * 2

	for _, alert := range a.Alerts {
		// check out values for the alert metric
		tenant = alert.Tenant
		oldState = alert.State
		newState = false

		// copy the metrics we need to check for this alert
		metrics := make([]string, len(alert.Metrics))
		copy(metrics, alert.Metrics)

		// add metrics from tags query
		if alert.Tags != "" {
			res, _ := a.Storage.GetItemList(tenant, storage.ParseTags(alert.Tags))
			for _, r := range res {
				metrics = append(metrics, r.ID)
			}
		}

		for _, metric := range metrics {
			rawData, _ := a.Storage.GetRawData(tenant, metric, end, start, limit, "DESC")

			// if we have new data check for alert status change
			// [ if no new data found, leave alert status un changed ]
			if len(rawData) > 0 {

				// update alert state and report to user if changed.
				newState = alert.alertState(rawData[0].Value)

				// if new state is true, remember the metrics
				if newState {
					// set triger values
					alert.TrigerMetric = metric
					alert.TrigerValue = rawData[0].Value
					alert.TrigerTimestamp = rawData[0].Timestamp

					// if this alert is on, just exit
					break
				}
			}
		}

		// if alert state changed, update state and triger event
		if newState != oldState {
			// update alert state
			alert.State = newState

			// triger event
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

// FilterAlerts return a list of alerts, filter by tenant, id and state
func (a *AlertRules) FilterAlerts(tenant string, id string, state string) []Alert {
	res := make([]Alert, 0)
	s := state == "T"

	for _, al := range a.Alerts {
		if al.Tenant == tenant && (id == "" || al.ID == id) && (state == "" || al.State == s) {
			res = append(res, *al)
		}
	}

	return res
}
