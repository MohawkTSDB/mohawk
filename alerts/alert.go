package alerts

import (
	"github.com/MohawkTSDB/mohawk/storage"
)

const (
	BETWEEN RangeIntervalType = iota
	HIGHER_THAN
	LOWER_THAN
)

type RangeIntervalType int

type Alert struct {
	Id      string         `mapstructure:"id"`
	Metric  string 		   `mapstructure:"metric"`
	Tenant  string         `mapstructure:"tenant"`
	State   bool		   `mapstructure:"state"`
	From    float64 	   `mapstructure:"from"`
	To      float64        `mapstructure:"to"`
	Type RangeIntervalType `mapstructure:"type"`
}

type Alerts struct{
	Backend storage.Backend `mapstrcuture: "storage"`
	Verbose bool			`mapstrcuture: "verbose"`
	Alerts  []*Alert        `mapstructure: "alerts"`
}

