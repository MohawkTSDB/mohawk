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
package sqlite

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/yaacov/mohawk/backend"
)

type Backend struct {
	dbDirName string
	tenant    map[string]*sql.DB
}

// Backend functions
// Required by backend interface

func (r Backend) Name() string {
	return "Backend-Sqlite3"
}

func (r *Backend) Open(options url.Values) {
	// get backend options
	r.dbDirName = options.Get("db-dirname")
	if r.dbDirName == "" {
		r.dbDirName = "./"
	}

	r.tenant = make(map[string]*sql.DB)
}

func (r Backend) GetTenants() []backend.Tenant {
	res := make([]backend.Tenant, 0)

	files, _ := ioutil.ReadDir(r.dbDirName)
	for _, f := range files {
		// take only sqlite db files as tenant names
		if p := strings.Split(f.Name(), "."); len(p) == 2 && p[1] == "db" {
			res = append(res, backend.Tenant{Id: p[0]})
		}
	}

	return res
}

func (r Backend) GetItemList(tenant string, tags map[string]string) []backend.Item {
	res := make([]backend.Item, 0)
	db, _ := r.GetTenant(tenant)

	// create one item per id
	sqlStmt := "select id from ids"
	rows, err := db.Query(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
	}
	defer rows.Close()
	for rows.Next() {
		var id string

		err = rows.Scan(&id)
		if err != nil {
			log.Printf("%q\n", err)
		}
		res = append(res, backend.Item{
			Id:   id,
			Type: "gauge",
			Tags: map[string]string{},
		})
	}
	err = rows.Err()
	if err != nil {
		log.Printf("%q\n", err)
	}

	// update item tags
	rows, err = db.Query("select id, tag, value from tags")
	if err != nil {
		log.Printf("%q\n", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var tag string
		var value string

		err = rows.Scan(&id, &tag, &value)
		if err != nil {
			log.Printf("%q\n", err)
		}
		res = r.UpdateTag(res, tenant, id, tag, value)
	}
	err = rows.Err()
	if err != nil {
		log.Printf("%q\n", err)
	}

	// filter using tags
	if len(tags) > 0 {
		for key, value := range tags {
			res = backend.FilterItems(res, func(i backend.Item) bool {
				r, _ := regexp.Compile("^" + value + "$")
				return r.MatchString(i.Tags[key])
			})
		}
	}

	return res
}

func (r Backend) GetRawData(tenant string, id string, end int64, start int64, limit int64, order string) []backend.DataItem {
	res := make([]backend.DataItem, 0)
	db, _ := r.GetTenant(tenant)

	// check if id exist
	if !r.IdExist(tenant, id) {
		return res
	}

	// id exist, get timestamp, value pairs
	sqlStmt := fmt.Sprintf(`select timestamp, value
		from '%s'
		where timestamp > %d and timestamp <= %d
		order by timestamp %s limit %d`,
		id, start, end, order, limit)
	rows, err := db.Query(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
	}
	defer rows.Close()
	for rows.Next() {
		var timestamp int64
		var value float64

		err = rows.Scan(&timestamp, &value)
		if err != nil {
			log.Printf("%q\n", err)
		}
		res = append(res, backend.DataItem{
			Timestamp: timestamp,
			Value:     value,
		})
	}
	err = rows.Err()
	if err != nil {
		log.Printf("%q\n", err)
	}

	return res
}

func (r Backend) GetStatData(tenant string, id string, end int64, start int64, limit int64, order string, bucketDuration int64) []backend.StatItem {
	var t int64
	res := make([]backend.StatItem, 0)
	db, _ := r.GetTenant(tenant)

	timeStep := bucketDuration * 1000
	startTime := int64(start/timeStep) * timeStep
	endTime := int64(end/timeStep) * timeStep

	// check if id exist
	if !r.IdExist(tenant, id) {
		return res
	}

	// id exist, get timestamp, value pairs
	sqlStmt := fmt.Sprintf(`select
		count(timestamp) as samples, cast((timestamp / %d) as integer) * %d as start, max(timestamp) as end,
		min(value) as min, max(value) as max, avg(value) as avg, sum(value) as sum
		from '%s'
		where timestamp >= %d and timestamp < %d
		group by start
		order by start ASC`,
		timeStep, timeStep, id, startTime, endTime)
	rows, err := db.Query(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
	}
	defer rows.Close()
	t = startTime
	for rows.Next() {
		var samples int64
		var startT int64
		var endT int64
		var min float64
		var max float64
		var avg float64
		var sum float64

		err = rows.Scan(&samples, &startT, &endT, &min, &max, &avg, &sum)
		if err != nil {
			log.Printf("%q\n", err)
		}

		// append missing
		for t < startT {
			res = append(res, backend.StatItem{
				Start:   t,
				End:     t + timeStep,
				Empty:   true,
				Samples: 0,
				Min:     0,
				Max:     0,
				Avg:     0,
				Median:  0,
				Sum:     0,
			})
			t += timeStep
		}

		// add data
		res = append(res, backend.StatItem{
			Start:   startT,
			End:     startT + timeStep,
			Empty:   false,
			Samples: samples,
			Min:     min,
			Max:     max,
			Avg:     avg,
			Median:  0,
			Sum:     sum,
		})
		t += timeStep
	}
	err = rows.Err()
	if err != nil {
		log.Printf("%q\n", err)
	}

	// append missing
	for t < endTime {
		res = append(res, backend.StatItem{
			Start:   t,
			End:     t + timeStep,
			Empty:   true,
			Samples: 0,
			Min:     0,
			Max:     0,
			Avg:     0,
			Median:  0,
			Sum:     0,
		})
		t += timeStep
	}

	return res
}

func (r Backend) PostRawData(tenant string, id string, t int64, v float64) bool {
	// check if id exist
	if !r.IdExist(tenant, id) {
		r.createId(tenant, id)
	}

	r.insertData(tenant, id, t, v)
	return true
}

func (r Backend) PutTags(tenant string, id string, tags map[string]string) bool {
	// check if id exist
	if !r.IdExist(tenant, id) {
		r.createId(tenant, id)
	}

	for k, v := range tags {
		r.insertTag(tenant, id, k, v)
	}
	return true
}

func (r Backend) DeleteData(tenant string, id string, end int64, start int64) bool {
	// check if id exist
	if r.IdExist(tenant, id) {
		r.deleteData(tenant, id, end, start)
		return true
	}

	return false
}

func (r Backend) DeleteTags(tenant string, id string, tags []string) bool {
	// check if id exist
	if r.IdExist(tenant, id) {
		for _, k := range tags {
			r.deleteTag(tenant, id, k)
		}
		return true
	}

	return false
}

// Helper functions
// Not required by backend interface

func (r *Backend) GetTenant(name string) (*sql.DB, error) {
	var filename string

	if tenant, ok := r.tenant[name]; ok {
		return tenant, nil
	}

	filename = fmt.Sprintf("%s/%s.db", r.dbDirName, name)

	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		log.Printf("%q\n", err)
		return nil, err
	}
	db.SetMaxOpenConns(1)

	sqlStmt := `
		create table if not exists ids (
			id    text,
			primary key (id));
		create table if not exists tags (
			id    text,
			tag   text,
			value text,
			primary key (id, tag));
		`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return db, err
	}

	r.tenant[name] = db
	return db, nil
}

func (r Backend) IdExist(tenant string, id string) bool {
	var _id string
	db, _ := r.GetTenant(tenant)

	sqlStmt := fmt.Sprintf("select id from ids where id='%s'", id)
	err := db.QueryRow(sqlStmt).Scan(&_id)
	return err != sql.ErrNoRows
}

func (r Backend) insertData(tenant string, id string, t int64, v float64) {
	db, _ := r.GetTenant(tenant)

	sqlStmt := fmt.Sprintf("insert into '%s' values (%d, %f)", id, t, v)
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
	}
}

func (r Backend) insertTag(tenant string, id string, k string, v string) {
	db, _ := r.GetTenant(tenant)

	sqlStmt := fmt.Sprintf("insert or replace into tags values ('%s', '%s', '%s')", id, k, v)
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
	}
}

func (r Backend) deleteData(tenant string, id string, end int64, start int64) {
	db, _ := r.GetTenant(tenant)

	sqlStmt := fmt.Sprintf("delete from '%s' where timestamp >= %d and timestamp < %d", id, start, end)
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
	}
}

func (r Backend) deleteTag(tenant string, id string, k string) {
	db, _ := r.GetTenant(tenant)

	sqlStmt := fmt.Sprintf("delete from tags where id='%s' and tag='%s'", id, k)
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
	}
}

func (r Backend) createId(tenant string, id string) bool {
	db, _ := r.GetTenant(tenant)

	sqlStmt := fmt.Sprintf("insert into ids values ('%s')", id)
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return false
	}

	sqlStmt = fmt.Sprintf(`
	create table if not exists '%s' (
		timestamp integer,
		value     numeric,
		primary key (timestamp));
	`, id)

	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return false
	}

	return true
}

func (r Backend) UpdateTag(items []backend.Item, tenant string, id string, tag string, value string) []backend.Item {
	// try to update tag if item exist
	for i, item := range items {
		if item.Id == id {
			items[i].Tags[tag] = value
			return items
		}
	}

	// if here we did not find a matching item
	items = append(items, backend.Item{
		Id:   id,
		Tags: map[string]string{tag: value},
	})

	return items
}
