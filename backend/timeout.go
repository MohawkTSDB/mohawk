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

type Timeout struct {
}

func (r Timeout) Name() string {
	return "Backend-Error"
}

func (r *Timeout) Open() {

}

func (r Timeout) GetItemList(tags map[string]string) []Item {
	res := make([]Item, 0)

	return res
}

func (r Timeout) GetRawData(id string, end int64, start int64, limit int64, order string) []DataItem {
	res := make([]DataItem, 0)

	return res
}

func (r Timeout) GetStatData(id string, end int64, start int64, limit int64, order string, bucketDuration int64) []StatItem {
	res := make([]StatItem, 0)

	return res
}

func (r Timeout) PostRawData(id string, t int64, v float64) bool {
	return false
}

func (r Timeout) PutTags(id string, tags map[string]string) bool {
	return false
}
