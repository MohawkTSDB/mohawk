package alerts

import (
	"github.com/MohawkTSDB/mohawk/storage"
	"github.com/MohawkTSDB/mohawk/storage/memory"
	"testing"
	"time"
)

var b storage.Backend

func TestAlerts_Init(t *testing.T) {
	var ti int64
	var va float64
	var state0 bool
	var state1 bool
	var state2 bool
	c := make(chan bool)

	// Testing with memory backend.
	b = &memory.Backend{}
	// Initialize backend object.
	b.Open(nil)

	// creating some alerts.
	l := []*Alert{
		{ID: "cpu usage too high",
			Tenant: "_ops",
			Type:   LOWER_THAN,
			To:     0.9,
			Metric: "cpu_usage"},
		{ID: "free memory too low ",
			Tenant: "_ops",
			Type:   HIGHER_THAN,
			From:   10000,
			Metric: "free_memory"},
		{ID: "free memory in between ",
			Tenant:"_ops",
			Type:   BETWEEN,
			From:   9000,
			To:     20000,
			Metric: "free_memory"},
	}

	// Create an alerts object with memory backend.
	alerts := Alerts{
		Alerts:  l,
		Backend: b,
		Verbose: true,
	}

	/////////
	// TEST 1:
	/////////

	// Create some fake data
	ti = int64(time.Now().UTC().Unix()*1000) - int64(30*60*1000)
	va = float64(9500) // Firing alert two
	b.PostRawData("_ops", "free_memory", ti, va)

	// run alerts worker in separate thread and push results to a channel:
	go func() {
		alerts.checkAlerts()
		for _, alert := range l {
			c <- alert.State
		}
	}()
	
	state0 = <- c
	state1 = <- c
	state2 = <- c

	// only alert two should fire!
	if !state1 || state0 || state2 {
		t.FailNow()
	}

	//////////
	// TEST 2:
	/////////

	// Create some more fake data
	ti = int64(time.Now().UTC().Unix()*1000) - int64(30*60*1000)
	va = float64(8500) // firing alert one and two
	b.PostRawData("_ops", "free_memory", ti, va)

	// run alerts worker in separate thread and push results to a channel:
	go func() {
		alerts.checkAlerts()
		for _, alert := range l {
			c <- alert.State
		}
	}()

	state0 = <- c
	state1 = <- c
	state2 = <- c

	// only alert two and one should fire!
	if !state1 || !state2 || state0 {
		t.FailNow()
	}

}
