package kludge

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

type DB struct {
	address string
	client  *http.Client
}

var versionRegexp = regexp.MustCompile("^kludge-\\d+\\.\\d|\\.\\d+$")
var ErrInvalidDatastore = fmt.Errorf("invalid datastore")

func Connect(addr string, client *http.Client) (db *DB, err error) {
	db = new(DB)
	if client == nil {
		db.client = http.DefaultClient
	}
	db.address = "http://" + addr

	if ver := db.Version(); ver == "" {
		err = ErrInvalidDatastore
	} else if !versionRegexp.MatchString(ver) {
		err = ErrInvalidDatastore
	}
	return
}

func (db *DB) Version() string {
	resp, err := db.client.Head(db.address + "/data")
	if err != nil {
		return ""
	}

	return resp.Header["X-Kludge-Version"][0]
}

func (db *DB) Get(key string) (value []byte, ok bool, err error) {
	url := db.address + "/data/" + key
	resp, err := db.client.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		ok = true
	}

	value, err = ioutil.ReadAll(resp.Body)
	return
}

func (db *DB) Set(key string, value []byte) (prev []byte, ok bool, err error) {
	url := db.address + "/data/" + key
	buf := bytes.NewBuffer(value)
	resp, err := db.client.Post(url, "application/json", buf)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		ok = true
	}

	prev, err = ioutil.ReadAll(resp.Body)
	return
}

func (db *DB) Del(key string) (prev []byte, ok bool, err error) {
	url := db.address + "/data/" + key
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return
	}
	resp, err := db.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		ok = true
	}

	prev, err = ioutil.ReadAll(resp.Body)
	return
}
