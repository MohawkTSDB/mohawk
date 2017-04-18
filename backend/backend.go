// Copyright 2016 Red Hat, Inc. and/or its affiliates
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

// Package backend
package backend

import (
	"net/url"
)

type Tenant struct {
	Id string `json:"id"`
}

type Item struct {
	Id   string            `json:"id"`
	Type string            `json:"type"`
	Tags map[string]string `json:"tags"`
}

type DataItem struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

type StatItem struct {
	Start   int64   `json:"start"`
	End     int64   `json:"end"`
	Empty   bool    `json:"empty"`
	Samples int64   `json:"samples"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Avg     float64 `json:"avg"`
	Median  float64 `json:"median"`
	Sum     float64 `json:"sum"`
}

type Backend interface {
	Name() string
	Open(options url.Values)
	GetItemList(tags map[string]string) []Item
	GetRawData(id string, end int64, start int64, limit int64, order string) []DataItem
	GetStatData(id string, end int64, start int64, limit int64, order string, bucketDuration int64) []StatItem
	PostRawData(id string, t int64, v float64) bool
	PutTags(id string, tags map[string]string) bool
	DeleteTags(id string, tags []string) bool
}

func FilterItems(vs []Item, f func(Item) bool) []Item {
	vsf := make([]Item, 0)

	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}
