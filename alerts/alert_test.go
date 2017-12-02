package alerts

import (
	"testing"
	"time"

	"github.com/MohawkTSDB/mohawk/storage/memory"
)

func TestAlerts_Init(test *testing.T) {
	var t int64
	var v float64

	// Testing with memory backend.
	b := &memory.Backend{}
	b.Open(nil)

	// creating some alerts.
	l := []*Alert{
		{
			ID:     "cpu usage too high",
			Tenant: "_ops",
			Metric: "cpu_usage",
			Type:   LOWER_THAN,
			To:     0.9,
		},
		{
			ID:     "free memory too low ",
			Tenant: "_ops",
			Metric: "free_memory",
			Type:   LOWER_THAN,
			From:   2000,
		},
		{
			ID:     "free memory in between ",
			Tenant: "_ops",
			Metric: "free_memory",
			Type:   OUTSIDE,
			From:   1000,
			To:     9000,
		},
		{
			ID:     "free memory in too high ",
			Tenant: "_ops",
			Metric: "free_memory",
			Type:   HIGHER_THAN,
			To:     4000,
		},
	}

	// Create an alerts object with memory backend.
	alerts := AlertRules{
		Alerts:  l,
		Backend: b,
		Verbose: true,
	}

	/////////
	// TEST 1:
	/////////

	// Create some fake data
	// Firing alert 1
	t = int64(time.Now().UTC().Unix()*1000) - int64(30*60*1000)
	v = float64(1500)
	b.PostRawData("_ops", "free_memory", t, v)

	// run alerts worker in separate thread and push results to a channel:
	alerts.checkAlerts()

	// only alert 1 should fire!
	if l[0].State || !l[1].State || l[2].State || l[3].State {
		test.Error("Fail test 1")
	}

	//////////
	// TEST 2:
	/////////

	// Create some more fake data
	// firing alerts 1 and 2
	t = int64(time.Now().UTC().Unix()*1000) - int64(30*60*1000)
	v = float64(500)
	b.PostRawData("_ops", "free_memory", t, v)

	// run alerts worker in separate thread and push results to a channel:
	alerts.checkAlerts()

	// only alerts 1 and 2 should fire!
	if l[0].State || !l[1].State || !l[2].State || l[3].State {
		test.Error("Fail test 2")
	}

	//////////
	// TEST 3:
	/////////

	// Create some more fake data
	// firing none
	t = int64(time.Now().UTC().Unix()*1000) - int64(30*60*1000)
	v = float64(2500)
	b.PostRawData("_ops", "free_memory", t, v)

	// run alerts worker in separate thread and push results to a channel:
	alerts.checkAlerts()

	// no alerts should fire!
	if l[0].State || l[1].State || l[2].State || l[3].State {
		test.Error("Fail test 3")
	}

	//////////
	// TEST 4:
	/////////

	// Create some more fake data
	// firing alert 3
	t = int64(time.Now().UTC().Unix()*1000) - int64(30*60*1000)
	v = float64(5000)
	b.PostRawData("_ops", "free_memory", t, v)

	// run alerts worker in separate thread and push results to a channel:
	alerts.checkAlerts()

	// alert 3 should fire
	if l[0].State || l[1].State || l[2].State || !l[3].State {
		test.Error("Fail test 3")
	}
}
