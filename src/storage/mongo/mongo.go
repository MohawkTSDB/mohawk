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

// Package mongo interface for mongo metric data storage
package mongo

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/MohawkTSDB/mohawk/src/storage"
)

type Storage struct {
	dbURL        string
	dbUsername   string
	dbPassword   string
	mongoSession *mgo.Session
}

// Storage functions
// Required by storage interface

// Name return a human readable storage name
func (r Storage) Name() string {
	return "Storage-Mongo"
}

// Help return a human readable storage help message
func (r Storage) Help() string {
	return `Mongo storage [mongo]:
	db-url   - comma separeted list of mongo servers.
	username - (optional) username for db access.
	password - (optional) password for db access.
	Examples:
		--options=db-url=42.153.3.25,42.153.3.26,42.153.3.27`
}

// Open storage
func (r *Storage) Open(options url.Values) {
	var err error

	// get storage options
	r.dbURL = options.Get("db-url")
	if r.dbURL == "" {
		r.dbURL = "127.0.0.1"
	}
	r.dbUsername = options.Get("username")
	r.dbPassword = options.Get("password")

	// We need this object to establish a session to our MongoDB.
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    strings.Split(r.dbURL, ","),
		Timeout:  10 * time.Second,
		Username: r.dbUsername,
		Password: r.dbPassword,
	}

	// Create a session which maintains a pool of socket connections
	// to our MongoDB.
	r.mongoSession, err = mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		panic(err)
	}

	r.mongoSession.SetMode(mgo.Monotonic, true)

	// log init arguments
	log.Printf("Start mongo storage:")
	log.Printf("  addrs: %+v", strings.Split(r.dbURL, ","))
}

func (r Storage) GetTenants() []storage.Tenant {
	res := make([]storage.Tenant, 0)

	// copy storage session
	sessionCopy := r.mongoSession.Copy()
	defer sessionCopy.Close()

	// return a list of tenants
	names, err := sessionCopy.DatabaseNames()
	if err != nil {
		log.Printf("%q\n", err)
		return res
	}
	for _, t := range names {
		if t != "admin" && t != "local" {
			res = append(res, storage.Tenant{ID: t})
		}
	}

	return res
}

func (r Storage) GetItemList(tenant string, tags map[string]string) []storage.Item {
	var query bson.M
	res := make([]storage.Item, 0)

	// copy storage session
	sessionCopy := r.mongoSession.Copy()
	defer sessionCopy.Close()

	c := sessionCopy.DB(tenant).C("ids")

	// Query taged items
	if len(tags) > 0 {
		query = bson.M{}

		for key, value := range tags {
			query["tags."+key] = bson.RegEx{"^" + value + "$", ""}
		}
	}

	err := c.Find(query).Sort("_id").All(&res)
	if err != nil {
		log.Printf("%q\n", err)
		return res
	}

	return res
}

func (r Storage) GetRawData(tenant string, id string, end int64, start int64, limit int64, order string) []storage.DataItem {
	var sort string
	res := make([]storage.DataItem, 0)

	// copy storage session
	sessionCopy := r.mongoSession.Copy()
	defer sessionCopy.Close()

	// order to sort
	if order == "DESC" {
		sort = "-timestamp"
	} else {
		sort = "timestamp"
	}

	c := sessionCopy.DB(tenant).C(id)

	// Query
	err := c.Find(bson.M{"timestamp": bson.M{"$gte": start, "$lt": end}}).Sort(sort).Limit(int(limit)).All(&res)
	if err != nil {
		log.Printf("%q\n", err)
		return res
	}

	return res
}

func (r Storage) GetStatData(tenant string, id string, end int64, start int64, limit int64, order string, bucketDuration int64) []storage.StatItem {
	var sort int
	res := make([]storage.StatItem, 0)

	// copy storage session
	sessionCopy := r.mongoSession.Copy()
	defer sessionCopy.Close()

	// order to sort
	if order == "DESC" {
		sort = -1
	} else {
		sort = 1
	}

	c := sessionCopy.DB(tenant).C(id)

	// Query
	err := c.Pipe(
		[]bson.M{
			{
				"$match": bson.M{"timestamp": bson.M{"$gte": start, "$lte": end}},
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"$trunc": bson.M{"$divide": []interface{}{"$timestamp", bucketDuration * 1000}},
					},
					"start": bson.M{"$first": bson.M{"$multiply": []interface{}{
						bson.M{"$trunc": bson.M{"$divide": []interface{}{
							"$timestamp",
							bucketDuration * 1000,
						}}},
						bucketDuration * 1000,
					}}},
					"end": bson.M{"$first": bson.M{"$multiply": []interface{}{
						bson.M{"$ceil": bson.M{"$divide": []interface{}{
							"$timestamp",
							bucketDuration * 1000,
						}}},
						bucketDuration * 1000,
					}}},
					"first":   bson.M{"$first": "$value"},
					"last":    bson.M{"$last": "$value"},
					"sum":     bson.M{"$sum": "$value"},
					"avg":     bson.M{"$avg": "$value"},
					"min":     bson.M{"$min": "$value"},
					"max":     bson.M{"$max": "$value"},
					"samples": bson.M{"$sum": 1},
				},
			},
			{
				"$sort": bson.M{"start": sort},
			},
			{
				"$limit": int(limit),
			},
		},
	).All(&res)
	if err != nil {
		log.Printf("%q\n", err)
		return res
	}
	return res
}

// unimplemented requests should fail silently

func (r Storage) PostRawData(tenant string, id string, t int64, v float64) bool {
	// check if id exist
	if !r.IDExist(tenant, id) {
		r.createID(tenant, id)
	}

	r.insertData(tenant, id, t, v)
	return true
}

func (r Storage) PutTags(tenant string, id string, tags map[string]string) bool {
	// check if id exist
	if !r.IDExist(tenant, id) {
		r.createID(tenant, id)
	}

	for k, v := range tags {
		r.insertTag(tenant, id, k, v)
	}
	return true
}

func (r Storage) DeleteData(tenant string, id string, end int64, start int64) bool {
	return true
}

func (r Storage) DeleteTags(tenant string, id string, tags []string) bool {
	return true
}

// Helper functions
// Not required by storage interface

func (r Storage) IDExist(tenant string, id string) bool {
	result := storage.Item{}

	// copy storage session
	sessionCopy := r.mongoSession.Copy()
	defer sessionCopy.Close()

	c := sessionCopy.DB(tenant).C("ids")

	err := c.Find(bson.M{"_id": id}).One(&result)
	return err == nil
}

func (r Storage) createID(tenant string, id string) bool {
	// copy storage session
	sessionCopy := r.mongoSession.Copy()
	defer sessionCopy.Close()

	c := sessionCopy.DB(tenant).C("ids")

	err := c.Insert(&storage.Item{ID: id, Type: "gauge", Tags: map[string]string{}, LastValues: []storage.DataItem{}})
	if err != nil {
		log.Printf("%q\n", err)
		return false
	}

	return true
}

func (r Storage) insertTag(tenant string, id string, k string, v string) {
	result := storage.Item{}

	// copy storage session
	sessionCopy := r.mongoSession.Copy()
	defer sessionCopy.Close()

	c := sessionCopy.DB(tenant).C("ids")

	// get current tags
	err := c.Find(bson.M{"_id": id}).One(&result)
	if err != nil {
		log.Printf("%q\n", err)
	}

	// Update
	result.Tags[k] = v
	err = c.Update(bson.M{"_id": id}, bson.M{"$set": bson.M{"tags": result.Tags}})
	if err != nil {
		log.Printf("%q\n", err)
	}
}

func (r Storage) insertData(tenant string, id string, t int64, v float64) {
	// copy storage session
	sessionCopy := r.mongoSession.Copy()
	defer sessionCopy.Close()

	c := sessionCopy.DB(tenant).C(id)
	err := c.Insert(&storage.DataItem{Timestamp: t, Value: v})

	if err != nil {
		log.Printf("%q\n", err)
	}
}
