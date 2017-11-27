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
	From float64		   `mapstructure:"from"`
	To   float64           `mapstructure:"to"`
	Type RangeIntervalType `mapstructure:"type"`
}

type Alert struct {
	Id      string         `mapstructure:"id"`
	Range   Range		   `mapstructure:"range"`
	State   bool		   `mapstructure:"state"`
}

type AlertList struct {
	Tenant  backend.Tenant `mapstructure:"tenant"`
	List []Alert 		   `mapstructure:"alert_list"`
}

type Alerts struct{
	Backend backend.Backend `mapstrcuture: "backend"`
	Verbose bool			`mapstrcuture: "verbose"`
	AlertLists []AlertList  `mapstructure: "alerts"`
}

