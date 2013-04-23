package kludge

import (
	"bytes"
	"encoding/json"
	"fmt"
        "github.com/gokyle/kludge/common"
	"io/ioutil"
	"net/http"
	"regexp"
)

// Type DataStore provides for datastore interaction.
type DataStore struct {
	address string
	client  *http.Client
}

var versionRegexp = regexp.MustCompile("^kludge-\\d+\\.\\d|\\.\\d+$")
var ErrInvalidDatastore = fmt.Errorf("invalid datastore")

// ClientVersion returns the client's version information.
func ClientVersion() string {
        return common.Version()
}

// Connect initialises a new DataStore value that will connect to the
// target datastore. It takes an address which should be an ip:port pointing
// to the front end, and a pointer to an http.Client. If the client is nil,
// the default client will be used.
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

// The Version method returns the datastore's version string.
func (ds *DataStore) Version() string {
	resp, err := ds.client.Head(ds.address + "/data")
	if err != nil {
		return ""
	}

	return resp.Header["X-Kludge-Version"][0]
}

// Get retrieves the value of a Unicode-encoded key from the datastore. It
// returns three arguments: the value of the key (if present), a boolean
// indicating whether the key is present in the datastore, and an error
// value storing any error that occurred retrieving the key's value.
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

// Set sets a new value for the key in the datastore. If the key is present,
// it is overwritten and the previous value returned. It returns three values:
// any previous value of the key, a boolean indicating whether the key was
// present in the datastore already, and an error containing any error that
// occurred setting the key.
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

// Del removes a key from the datastore. It returns three values: the value of
// the key if it is present, a boolean indicating whether the key was present
// in the database, and any error that occurred while deleting the key. The
// boolean will be true if the key was in the database and removed.
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

// List returns a slice of all the keys in the datastore as Unicode strings.
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
