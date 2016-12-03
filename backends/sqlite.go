package backend

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Sqlite struct {
	db *sql.DB
}

// Backend functions
// Required by backend interface

func (r Sqlite) Name() string {
	return "Backend-Sqlite"
}

func (r *Sqlite) Open() {
	var err error

	r.db, err = sql.Open("sqlite3", "./server.db")
	if err != nil {
		log.Fatal(err)
	}

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
	_, err = r.db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
}

func (r Sqlite) GetItemList(tags map[string]string) []Item {
	res := make([]Item, 0)

	// create one item per id
	rows, err := r.db.Query("select id from ids")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id string

		err = rows.Scan(&id)
		if err != nil {
			log.Fatal(err)
		}
		res = append(res, Item{
			Id:   id,
			Type: "gauge",
			Tags: map[string]string{},
		})
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	// update item tags
	rows, err = r.db.Query("select id, tag, value from tags")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var tag string
		var value string

		err = rows.Scan(&id, &tag, &value)
		if err != nil {
			log.Fatal(err)
		}
		res = r.UpdateTag(res, id, tag, value)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	// filter using tags
	// FIXME: we should do this in the sql
	if len(tags) > 0 {
		for key, value := range tags {
			res = FilterItems(res, func(i Item) bool { return i.Tags[key] == value })
		}
	}

	return res
}

func (r Sqlite) GetRawData(id string, end int64, start int64, limit int64, order string) []DataItem {
	res := make([]DataItem, 0)

	// check if id exist
	if !r.IdExist(id) {
		return res
	}

	// id exist, get timestamp, value pairs
	sqlStmt := fmt.Sprintf(`select timestamp, value
		from %s
		where timestamp > %d and timestamp <= %d
		order by timestamp %s limit %d`,
		id, start, end, order, limit)
	rows, err := r.db.Query(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var timestamp int64
		var value float64

		err = rows.Scan(&timestamp, &value)
		if err != nil {
			log.Fatal(err)
		}
		res = append(res, DataItem{
			Timestamp: timestamp,
			Value:     value,
		})
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return res
}

func (r Sqlite) GetStatData(id string, end int64, start int64, limit int64, order string, bucketDuration int64) []StatItem {
	var t int64
	res := make([]StatItem, 0)

	// check if id exist
	if !r.IdExist(id) {
		return res
	}

	// id exist, get timestamp, value pairs
	sqlStmt := fmt.Sprintf(`select
		count(timestamp) as samples, cast((timestamp / %d) as integer) * %d as start, max(timestamp) as end,
		min(value) as min, max(value) as max, avg(value) as avg, sum(value) as sum
		from %s
		where timestamp > %d and timestamp <= %d
		group by start
		order by start %s limit %d`,
		bucketDuration*1000, bucketDuration*1000, id, start, end, order, limit)
	rows, err := r.db.Query(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	t = int64(start/(bucketDuration*1000)) * (bucketDuration * 1000)
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
			log.Fatal(err)
		}

		// append missing
		for t < (startT - bucketDuration*1000) {
			res = append(res, StatItem{
				Start:   t,
				End:     t + bucketDuration*1000,
				Empty:   true,
				Samples: 0,
				Min:     0,
				Max:     0,
				Avg:     0,
				Median:  0,
				Sum:     0,
			})
			t += bucketDuration * 1000
		}

		// add data
		res = append(res, StatItem{
			Start:   startT,
			End:     startT + bucketDuration*1000,
			Empty:   false,
			Samples: samples,
			Min:     min,
			Max:     max,
			Avg:     avg,
			Median:  0,
			Sum:     sum,
		})
		t += bucketDuration * 1000
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	// append missing
	for t < (end - bucketDuration*1000) {
		res = append(res, StatItem{
			Start:   t,
			End:     t + bucketDuration*1000,
			Empty:   true,
			Samples: 0,
			Min:     0,
			Max:     0,
			Avg:     0,
			Median:  0,
			Sum:     0,
		})
		t += bucketDuration * 1000
	}

	return res
}

func (r Sqlite) PostRawData(id string, t int64, v float64) bool {
	// check if id exist
	if !r.IdExist(id) {
		r.createId(id)
	}

	r.insertData(id, t, v)
	return true
}

func (r Sqlite) PutTags(id string, tags map[string]string) bool {
	// check if id exist
	if !r.IdExist(id) {
		r.createId(id)
	}

	for k, v := range tags {
		r.insertTag(id, k, v)
	}
	return true
}

// Helper functions
// Not required by backend interface

func (r Sqlite) IdExist(id string) bool {
	var _id string
	sqlStmt := fmt.Sprintf("select id from ids where id='%s'", id)
	err := r.db.QueryRow(sqlStmt).Scan(&_id)
	return err != sql.ErrNoRows
}

func (r Sqlite) insertData(id string, t int64, v float64) {
	sqlStmt := fmt.Sprintf("insert into %s values (%d, %f)", id, t, v)
	_, err := r.db.Exec(sqlStmt)
	if err != nil {
		log.Fatal("%q: %s\n", err, sqlStmt)
	}
}

func (r Sqlite) insertTag(id string, k string, v string) {
	sqlStmt := fmt.Sprintf("insert or replace into tags values ('%s', '%s', '%s')", id, k, v)
	_, err := r.db.Exec(sqlStmt)
	if err != nil {
		log.Fatal("%q: %s\n", err, sqlStmt)
	}
}

func (r Sqlite) createId(id string) bool {
	sqlStmt := fmt.Sprintf("insert into ids values ('%s')", id)
	_, err := r.db.Exec(sqlStmt)
	if err != nil {
		log.Fatal("%q: %s\n", err, sqlStmt)
		return false
	}

	sqlStmt = fmt.Sprintf(`
	create table if not exists %s (
		timestamp integer,
		value     numeric,
		primary key (timestamp));
	`, id)

	_, err = r.db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return false
	}

	return true
}

func (r Sqlite) UpdateTag(items []Item, id string, tag string, value string) []Item {
	// try to update tag if item exist
	for i, item := range items {
		if item.Id == id {
			items[i].Tags[tag] = value
			return items
		}
	}

	// if here we did not find a matching item
	items = append(items, Item{
		Id:   id,
		Tags: map[string]string{tag: value},
	})

	return items
}
