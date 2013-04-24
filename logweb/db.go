package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

type logEntry struct {
	Node string
	Time string
	Msg  string
}

type responseTime struct {
	Node string
	Get  string
	Set  string
	Del  string
	Lst  string
}

func responseNodes() (node_list []string, err error) {
	node_list = make([]string, 0)
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return
	}
	defer db.Close()

	rows, err := db.Query("select distinct node from response_time")
	if err != nil {
		return
	}

	var node string
	for rows.Next() {
		err = rows.Scan(&node)
		if err != nil {
			rows.Close()
			return
		}
		node_list = append(node_list, node)
	}
	return
}

func getNodeAverages(node string) (resp *responseTime, err error) {
	responses := make(map[string]string, 0)
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return
	}
	defer db.Close()

	rows, err := db.Query(`
            select operation, avg(microsec) from response_time
            where node=?
            group by operation`, node)

	var operation string
	var rtime float64
	for rows.Next() {
		err = rows.Scan(&operation, &rtime)
		if err != nil {
			rows.Close()
			return
		}
		responses[operation] = fmt.Sprintf("%.2f", rtime)
	}
	resp = new(responseTime)
	resp.Node = node
	resp.Get = responses["GET"]
	resp.Set = responses["SET"]
	resp.Del = responses["DEL"]
	resp.Lst = responses["LST"]
	return
}

func getLastHour() (entries []*logEntry, err error) {
	hour, err := time.ParseDuration("-1h")
	if err != nil {
		return
	}

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return
	}
	defer db.Close()

	var (
		id        int64
		node      string
		timestamp int64
		msg       string
	)

	entries = make([]*logEntry, 0)
	whence := time.Now().Add(hour).Unix()
	rows, err := db.Query(`select * from entries 
            where timestamp > ? 
            order by timestamp desc`, whence)
	if err != nil {
		return
	}

	for rows.Next() {
		err = rows.Scan(&id, &node, &timestamp, &msg)
		if err != nil {
			return
		}
		le := new(logEntry)
		le.Node = node
		le.Time, err = utcTimestampToString(timestamp)
		le.Msg = msg
		entries = append(entries, le)
	}
	return
}

func utcTimestampToString(timestamp int64) (timeString string, err error) {
	ltime := time.Unix(timestamp, 0)
	utc, err := time.LoadLocation("UTC")
	if err != nil {
		return
	}

	utcTime := time.Date(ltime.Year(), ltime.Month(), ltime.Day(),
		ltime.Hour(), ltime.Minute(), ltime.Second(),
		0, utc)
	timeString = utcTime.Local().String()
	return
}
