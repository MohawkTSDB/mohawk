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
	id      string
	r       Range
	state   bool
}

type AlertList struct {
	tenant  backend.Tenant
	list []Alert `mapstructure:"alert_list`
}

type Alerts struct{
	Backend backend.Backend
	Verbose bool
	AlertLists []AlertList `mapstructure:"alerts`
}

