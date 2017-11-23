package alerts

import (
	"github.com/MohawkTSDB/mohawk/backend"
)

const (
	BETWEEN RangeIntervalType = iota
	HIGHER_THAN
	LOWER_THAN
)

type RangeIntervalType int

type Range struct {
	from float64
	to   float64
	t    RangeIntervalType
}

type Alert struct {
	tenant  backend.Tenant
	id      string
	r       Range
	state   bool
}

type Alerts struct {
	Verbose bool
	Backend backend.Backend
	AlertsList []Alert
}

