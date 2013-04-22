package kludge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

// Type DB provides for keystore interaction.
type DataStore struct {
	address string
	client  *http.Client
}

var versionRegexp = regexp.MustCompile("^kludge-\\d+\\.\\d|\\.\\d+$")
var ErrInvalidDatastore = fmt.Errorf("invalid datastore")

func Connect(addr string, client *http.Client) (ds *DataStore, err error) {
	ds = new(DataStore)
	if client == nil {
		ds.client = http.DefaultClient
	}
	ds.address = "http://" + addr

	if ver := ds.Version(); ver == "" {
		err = ErrInvalidDatastore
	} else if !versionRegexp.MatchString(ver) {
		err = ErrInvalidDatastore
	}
	return
}

func (ds *DataStore) Version() string {
	resp, err := ds.client.Head(ds.address + "/data")
	if err != nil {
		return ""
	}

	return resp.Header["X-Kludge-Version"][0]
}

func (ds *DataStore) Get(key string) (value []byte, ok bool, err error) {
	url := ds.address + "/data/" + key
	resp, err := ds.client.Get(url)
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

func (ds *DataStore) Set(key string, value []byte) (prev []byte, ok bool, err error) {
	url := ds.address + "/data/" + key
	buf := bytes.NewBuffer(value)
	resp, err := ds.client.Post(url, "application/json", buf)
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

func (ds *DataStore) Del(key string) (prev []byte, ok bool, err error) {
	url := ds.address + "/data/" + key
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return
	}
	resp, err := ds.client.Do(req)
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

func (ds *DataStore) List() (keys []string, err error) {
	url := ds.address + "/data"
	resp, err := ds.client.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	keys = make([]string, 0)
	err = json.Unmarshal(body, &keys)
	return
}
