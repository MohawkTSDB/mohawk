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

// Package storage interface for metric data storage
package storage

import (
	"net/url"
)

// Tenant just the tenant name
type Tenant struct {
	ID string `json:"id"`
}

// Item one metric item
type Item struct {
	ID         string            `json:"id" bson:"_id"`
	Type       string            `json:"type" bson:"type"`
	Tags       map[string]string `json:"tags" bson:"tags"`
	LastValues []DataItem        `json:"data,omitempty" bson:"data,omitempty"`
}

// DataItem one metric data point
type DataItem struct {
	Timestamp int64   `json:"timestamp" bson:"timestamp"`
	Value     float64 `json:"value" bson:"value"`
}

// StatItem one statistics data point
type StatItem struct {
	Start   int64   `json:"start"`
	End     int64   `json:"end"`
	Empty   bool    `json:"empty"`
	Samples int64   `json:"samples,omitempty"`
	Min     float64 `json:"min,omitempty"`
	Max     float64 `json:"max,omitempty"`
	First   float64 `json:"first,omitempty"`
	Last    float64 `json:"last,omitempty"`
	Avg     float64 `json:"avg,omitempty"`
	Median  float64 `json:"median,omitempty"`
	Std     float64 `json:"std,omitempty"`
	Sum     float64 `json:"sum,omitempty"`
}

// Storage metric data interface
type Storage interface {
	Name() string
	Open(options url.Values)
	GetTenants() []Tenant
	GetItemList(tenant string, tags map[string]string) []Item
	GetRawData(tenant string, id string, end int64, start int64, limit int64, order string) []DataItem
	GetStatData(tenant string, id string, end int64, start int64, limit int64, order string, bucketDuration int64) []StatItem
	PostRawData(tenant string, id string, t int64, v float64) bool
	PutTags(tenant string, id string, tags map[string]string) bool
	DeleteData(tenant string, id string, end int64, start int64) bool
	DeleteTags(tenant string, id string, tags []string) bool
}
