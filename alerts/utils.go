package alerts

import (
	"time"
	"fmt"
	"encoding/json"
)

func (alert *Alert) updateAlertState(value float64) {
	switch alert.Type {
	case BETWEEN:
		alert.State = value <= alert.From || value > alert.To
	case LOWER_THAN:
		alert.State = value > alert.To
	case HIGHER_THAN:
		alert.State = value < alert.From
	}
}

func (a *Alerts) Open() {
	// if user omit the tenant field in the alerts config, fallback to default
	// tenant
	for _, alert := range a.Alerts {
		// fall back to _ops
		if alert.Tenant == "" {
			alert.Tenant = "_ops"
		}
	}

	// start a maintenance worker that will check for alerts periodically.
	go a.maintenance()
}

func (a *Alerts) maintenance() {
	c := time.Tick(time.Second * 10)

	// once a minute check for alerts in data
	for range c {
		fmt.Printf("alert check: start\n")
		a.checkAlerts()
	}
}

func (a *Alerts) checkAlerts() {
	var end    int64
	var start  int64
	var tenant string
	var metric string
	var oldState bool

	for _, alert := range a.Alerts {
		// Get each tenants item list
		end = int64(time.Now().UTC().Unix() * 1000)
		start = end - 60*60*1000 // one hour ago

		tenant = alert.Tenant
		metric = alert.Metric
		rawData := a.Backend.GetRawData(tenant, metric, end, start, 1, "ASC")
		
		// check for alert status change
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


