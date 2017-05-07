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
package timeout

import (
	"net/url"

	"github.com/yaacov/mohawk/backend"
)

type Backend struct {
}

func (r Backend) Name() string {
	return "Backend-TimeoutError"
}

func (r *Backend) Open(options url.Values) {
}

func (r Backend) GetTenants() []backend.Tenant {
	var res []backend.Tenant

	res = append(res, backend.Tenant{Id: "_ops"})
	return res
}

func (r Backend) GetItemList(tenant string, tags map[string]string) []backend.Item {
	var res []backend.Item

	return res
}

func (r Backend) GetRawData(tenant string, id string, end int64, start int64, limit int64, order string) []backend.DataItem {
	var res []backend.DataItem

	return res
}

func (r Backend) GetStatData(tenant string, id string, end int64, start int64, limit int64, order string, bucketDuration int64) []backend.StatItem {
	var res []backend.StatItem

	return res
}

func (r Backend) PostRawData(tenant string, id string, t int64, v float64) bool {
	return false
}

func (r Backend) PutTags(tenant string, id string, tags map[string]string) bool {
	return false
}

func (r Backend) DeleteData(tenant string, id string, end int64, start int64) bool {
	return false
}

func (r Backend) DeleteTags(tenant string, id string, tags []string) bool {
	return false
}
