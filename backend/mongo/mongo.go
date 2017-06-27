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
package mongo

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"net/url"
	"regexp"
	"time"

	"github.com/yaacov/mohawk/backend"
)

type Backend struct {
	dbURL        string
	mongoSession *mgo.Session
}

// Backend functions
// Required by backend interface

func (r Backend) Name() string {
	return "Backend-Mongo"
}

func (r *Backend) Open(options url.Values) {
	var err error

	// get backend options
	r.dbURL = options.Get("db-url")
	if r.dbURL == "" {
		r.dbURL = "127.0.0.1"
	}

	// We need this object to establish a session to our MongoDB.
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{r.dbURL},
		Timeout:  10 * time.Second,
		Username: "",
		Password: "",
	}

	// Create a session which maintains a pool of socket connections
	// to our MongoDB.
	r.mongoSession, err = mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		panic(err)
	}

	r.mongoSession.SetMode(mgo.Monotonic, true)
}

func (r Backend) GetTenants() []backend.Tenant {
	res := make([]backend.Tenant, 0)

	// copy backend session
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
			res = append(res, backend.Tenant{Id: t})
		}
	}

	return res
}

func (r Backend) GetItemList(tenant string, tags map[string]string) []backend.Item {
	res := make([]backend.Item, 0)

	// copy backend session
	sessionCopy := r.mongoSession.Copy()
	defer sessionCopy.Close()

	c := sessionCopy.DB(tenant).C("ids")

	// Query All
	// TODO: find only taged items
	findTags := make([]bson.M, 0)

	err := c.Find(bson.M{"$and": findTags}).Sort("_id").All(&res)
	if err != nil {
		log.Printf("%q\n", err)
		return res
	}

	// filter using tags
	// 	if we have a list of _all_ items, we need to filter them by tags
	// 	if the list is already filtered, we do not need to re-filter it
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
	var sort string
	res := make([]backend.DataItem, 0)

	// copy backend session
	sessionCopy := r.mongoSession.Copy()
	defer sessionCopy.Close()

	// order to sort
	if order == "DESC" {
		sort = "timestamp"
	} else {
		sort = "-timestamp"
	}

	c := sessionCopy.DB(tenant).C(id)

	// Query
	err := c.Find(bson.M{"timestamp": bson.M{"$gte": start, "$lte": end}}).Sort(sort).Limit(int(limit)).All(&res)
	if err != nil {
		log.Printf("%q\n", err)
		return res
	}

	return res
}

func (r Backend) GetStatData(tenant string, id string, end int64, start int64, limit int64, order string, bucketDuration int64) []backend.StatItem {
	var sort int
	res := make([]backend.StatItem, 0)

	// copy backend session
	sessionCopy := r.mongoSession.Copy()
	defer sessionCopy.Close()

	// order to sort
	if order == "DESC" {
		sort = 1
	} else {
		sort = -1
	}

	c := sessionCopy.DB(tenant).C(id)

	// Query
	err := c.Pipe(
		[]bson.M{
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
				"$sort":  bson.M{"$start": sort},
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
	return true
}

func (r Backend) DeleteTags(tenant string, id string, tags []string) bool {
	return true
}

// Helper functions
// Not required by backend interface

func (r Backend) IdExist(tenant string, id string) bool {
	result := backend.Item{}

	// copy backend session
	sessionCopy := r.mongoSession.Copy()
	defer sessionCopy.Close()

	c := sessionCopy.DB(tenant).C("ids")

	err := c.Find(bson.M{"_id": id}).One(&result)
	return err == nil
}

func (r Backend) createId(tenant string, id string) bool {
	// copy backend session
	sessionCopy := r.mongoSession.Copy()
	defer sessionCopy.Close()

	c := sessionCopy.DB(tenant).C("ids")

	// TODO: check id name is not "ids" :-)
	// TODO: check tenant name is not "admin" or "local" :-)
	err := c.Insert(&backend.Item{Id: id, Type: "gauge", Tags: map[string]string{}, LastValues: []backend.DataItem{}})
	if err != nil {
		log.Printf("%q\n", err)
		return false
	}

	return true
}

func (r Backend) insertTag(tenant string, id string, k string, v string) {
	result := backend.Item{}

	// copy backend session
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

func (r Backend) insertData(tenant string, id string, t int64, v float64) {
	// copy backend session
	sessionCopy := r.mongoSession.Copy()
	defer sessionCopy.Close()

	c := sessionCopy.DB(tenant).C(id)
	err := c.Insert(&backend.DataItem{Timestamp: t, Value: v})

	if err != nil {
		log.Printf("%q\n", err)
	}
}
