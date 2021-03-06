// Copyright 2016,2017,2018 Yaacov Zamir <kobi.zamir@gmail.com>
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

// Package sqlite interface for sqlite metric data storage
package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"regexp"
	"strings"

	"github.com/MohawkTSDB/mohawk/src/storage"
	// go-sqlite3 is used by the database/sql package
	_ "github.com/mattn/go-sqlite3"
)

// errBadMetricID a new error with bad metrics id message
var errBadMetricID = errors.New("sqlite: Bad metrics ID")

type Storage struct {
	dbDirName string
	tenant    map[string]*sql.DB
}

// Storage functions
// Required by storage interface

// Name return a human readable storage name
func (r Storage) Name() string {
	return "Storage-Sqlite3"
}

// Help return a human readable storage help message
func (r Storage) Help() string {
	return `Mongo storage [mongo]:
	db-dirname - a directory for sqlite db file storage.
	Examples:
		--options=db-dirname=/data`
}

// Open storage
func (r *Storage) Open(options url.Values) {
	// get storage options
	r.dbDirName = options.Get("db-dirname")
	if r.dbDirName == "" {
		r.dbDirName = "."
	}

	r.tenant = make(map[string]*sql.DB)

	// log init arguments
	log.Printf("Start sqlite storage:")
	log.Printf("  db dirname: %+v", r.dbDirName)
}

func (r Storage) GetTenants() ([]storage.Tenant, error) {
	res := make([]storage.Tenant, 0)

	files, _ := ioutil.ReadDir(r.dbDirName)
	for _, f := range files {
		// take only sqlite db files as tenant names
		if p := strings.Split(f.Name(), "."); len(p) == 2 && p[1] == "db" {
			res = append(res, storage.Tenant{ID: p[0]})
		}
	}

	return res, nil
}

func (r Storage) GetItemList(tenant string, tags map[string]string) ([]storage.Item, error) {
	res := make([]storage.Item, 0)
	db, _ := r.getTenant(tenant)

	// create one item per id
	sqlStmt := "select id from ids"
	rows, err := db.Query(sqlStmt)
	if err != nil {
		return res, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string

		err = rows.Scan(&id)
		if err != nil {
			return res, err
		}
		res = append(res, storage.Item{
			ID:   id,
			Type: "gauge",
			Tags: map[string]string{},
		})
	}
	err = rows.Err()
	if err != nil {
		return res, err
	}

	// update item tags
	rows, err = db.Query("select id, tag, value from tags")
	if err != nil {
		return res, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var tag string
		var value string

		err = rows.Scan(&id, &tag, &value)
		if err != nil {
			return res, err
		}
		res = r.updateTag(res, tenant, id, tag, value)
	}
	err = rows.Err()
	if err != nil {
		return res, err
	}

	// filter using tags
	if len(tags) > 0 {
		for key, value := range tags {
			res = storage.FilterItems(res, func(i storage.Item) bool {
				r, _ := regexp.Compile("^" + value + "$")
				return r.MatchString(i.Tags[key])
			})
		}
	}

	return res, err
}

func (r Storage) GetRawData(tenant string, id string, end int64, start int64, limit int64, order string) ([]storage.DataItem, error) {
	res := make([]storage.DataItem, 0)
	db, _ := r.getTenant(tenant)

	// check if id exist
	if !r.IDExist(tenant, id) {
		return res, errBadMetricID
	}

	// id exist, get timestamp, value pairs
	sqlStmt := fmt.Sprintf(`select timestamp, value
		from '%s'
		where timestamp >= %d and timestamp < %d
		order by timestamp %s limit %d`,
		id, start, end, order, limit)
	rows, err := db.Query(sqlStmt)
	if err != nil {
		return res, err
	}
	defer rows.Close()
	for rows.Next() {
		var timestamp int64
		var value float64

		err = rows.Scan(&timestamp, &value)
		if err != nil {
			return res, err
		}
		res = append(res, storage.DataItem{
			Timestamp: timestamp,
			Value:     value,
		})
	}
	err = rows.Err()

	return res, err
}

func (r Storage) GetStatData(tenant string, id string, end int64, start int64, limit int64, order string, bucketDuration int64) ([]storage.StatItem, error) {
	var samples int64
	var startT int64
	var endT int64
	var min float64
	var max float64
	var avg float64
	var sum float64

	count := int64(0)
	res := make([]storage.StatItem, 0)
	db, _ := r.getTenant(tenant)

	timeStep := bucketDuration * 1000
	startTime := int64(start/timeStep) * timeStep
	endTime := int64(1+end/timeStep) * timeStep

	// check if id exist
	if !r.IDExist(tenant, id) {
		return res, errBadMetricID
	}

	// id exist, get timestamp, value pairs
	sqlStmt := fmt.Sprintf(`select
		count(timestamp) as samples, cast((timestamp / %d) as integer) * %d as start, max(timestamp) as end,
		min(value) as min, max(value) as max, avg(value) as avg, sum(value) as sum
		from '%s'
		where timestamp >= %d and timestamp < %d
		group by start
		order by start %s`,
		timeStep, timeStep, id, startTime, endTime, order)
	rows, err := db.Query(sqlStmt)
	if err != nil {
		return res, err
	}
	defer rows.Close()

	for rows.Next() && count < limit {
		err = rows.Scan(&samples, &startT, &endT, &min, &max, &avg, &sum)
		if err != nil {
			return res, err
		}

		// add data
		count++
		res = append(res, storage.StatItem{
			Start:   startT,
			End:     startT + timeStep,
			Empty:   false,
			Samples: samples,
			Min:     min,
			Max:     max,
			Avg:     avg,
			Sum:     sum,
		})
	}
	err = rows.Err()

	return res, err
}

// PostRawData handle posting data to db
func (r Storage) PostRawData(tenant string, id string, t int64, v float64) error {
	// check if id exist
	if !r.IDExist(tenant, id) {
		if err := r.createID(tenant, id); err != nil {
			return err
		}
	}

	err := r.insertData(tenant, id, t, v)
	return err
}

// PutTags handle posting tags to db
func (r Storage) PutTags(tenant string, id string, tags map[string]string) error {
	// check if id exist
	if !r.IDExist(tenant, id) {
		if err := r.createID(tenant, id); err != nil {
			return err
		}
	}

	for k, v := range tags {
		if err := r.insertTag(tenant, id, k, v); err != nil {
			return err
		}
	}
	return nil
}

// DeleteData handle delete data fron db
func (r Storage) DeleteData(tenant string, id string, end int64, start int64) error {
	// check if id exist
	if r.IDExist(tenant, id) {
		err := r.deleteData(tenant, id, end, start)
		return err
	}

	return errors.New("slite: ID not found")
}

// DeleteTags handle delete tags fron db
func (r Storage) DeleteTags(tenant string, id string, tags []string) error {
	// check if id exist
	if r.IDExist(tenant, id) {
		for _, k := range tags {
			if err := r.deleteTag(tenant, id, k); err != nil {
				return err
			}
		}
		return nil
	}

	return errors.New("sqlite: ID not found")
}

// Helper functions
// Not required by storage interface

func (r *Storage) getTenant(name string) (*sql.DB, error) {
	var filename string

	if tenant, ok := r.tenant[name]; ok {
		return tenant, nil
	}

	filename = fmt.Sprintf("%s/%s.db", r.dbDirName, name)

	db, err := sql.Open("sqlite3", filename)
	if err != nil {
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
	if err == nil {
		r.tenant[name] = db
	}

	return db, err
}

func (r Storage) IDExist(tenant string, id string) bool {
	var _id string
	db, err := r.getTenant(tenant)
	if err != nil {
		return false
	}

	sqlStmt := fmt.Sprintf("select id from ids where id='%s'", id)
	err = db.QueryRow(sqlStmt).Scan(&_id)
	return err != sql.ErrNoRows
}

func (r Storage) insertData(tenant string, id string, t int64, v float64) error {
	db, err := r.getTenant(tenant)
	if err != nil {
		return err
	}

	sqlStmt := fmt.Sprintf("insert into '%s' values (%d, %f)", id, t, v)
	_, err = db.Exec(sqlStmt)

	return err
}

func (r Storage) insertTag(tenant string, id string, k string, v string) error {
	db, err := r.getTenant(tenant)
	if err != nil {
		return err
	}
	sqlStmt := fmt.Sprintf("insert or replace into tags values ('%s', '%s', '%s')", id, k, v)
	_, err = db.Exec(sqlStmt)

	return err
}

func (r Storage) deleteData(tenant string, id string, end int64, start int64) error {
	db, err := r.getTenant(tenant)
	if err != nil {
		return err
	}

	sqlStmt := fmt.Sprintf("delete from '%s' where timestamp >= %d and timestamp < %d", id, start, end)
	_, err = db.Exec(sqlStmt)

	return err
}

func (r Storage) deleteTag(tenant string, id string, k string) error {
	db, err := r.getTenant(tenant)
	if err != nil {
		return err
	}

	sqlStmt := fmt.Sprintf("delete from tags where id='%s' and tag='%s'", id, k)
	_, err = db.Exec(sqlStmt)

	return err
}

func (r Storage) createID(tenant string, id string) error {
	db, err := r.getTenant(tenant)
	if err != nil {
		return err
	}

	sqlStmt := fmt.Sprintf("insert into ids values ('%s')", id)
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return err
	}

	sqlStmt = fmt.Sprintf(`
	create table if not exists '%s' (
		timestamp integer,
		value     numeric,
		primary key (timestamp));
	`, id)

	_, err = db.Exec(sqlStmt)

	return err
}

func (r Storage) updateTag(items []storage.Item, tenant string, id string, tag string, value string) []storage.Item {
	// try to update tag if item exist
	for i, item := range items {
		if item.ID == id {
			items[i].Tags[tag] = value
			return items
		}
	}

	// if here we did not find a matching item
	items = append(items, storage.Item{
		ID:   id,
		Tags: map[string]string{tag: value},
	})

	return items
}
