package main

import (
	"github.com/gokyle/kludge/common"
	"io"
	"log"
	"net/http"
	"regexp"
)

var keyIDRegexp = regexp.MustCompile("^/data/(.+)$")

func ServerError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}

func NotImplemented(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	msg := "Method " + r.Method + " not implemented."
	w.Write([]byte(msg))
}

func VersionHeader(w http.ResponseWriter) {
	version := common.Version()
	w.Header().Add("X-Kludge-Version", version)
}

func KeyID(r *http.Request) string {
	return keyIDRegexp.ReplaceAllString(r.URL.Path, "$1")
}

func ListKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := listKeys()
	if err != nil {
		ServerError(w, err)
		return
	}
	w.Write(keys)
}

func GetKey(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/data" || r.URL.Path == "/data/" {
		ListKeys(w, r)
		return
	}
	key := KeyID(r)
	body, ok, err := getKey(key)
	if err != nil {
		ServerError(w, err)
		return
	}
	if !ok {
		w.WriteHeader(http.StatusNotFound)
	}
	w.Write(body)
}

func DelKey(w http.ResponseWriter, r *http.Request) {
	body, ok, err := delKey(KeyID(r))
	if err != nil {
		ServerError(w, err)
		return
	}
	if !ok {
		w.WriteHeader(http.StatusNotFound)
	}
	w.Write(body)
}

func SetKey(w http.ResponseWriter, r *http.Request) {
	key := KeyID(r)
	defer r.Body.Close()

	var value []byte
	if r.ContentLength > 0 {
		value = make([]byte, r.ContentLength)
	} else {
		value = make([]byte, 0)
	}
	_, err := io.ReadFull(r.Body, value)
	if err != nil {
		log.Printf("request for %s failed: %s", r.URL.String(),
			err.Error())
		ServerError(w, err)
		return
	}
	body, ok, err := setKey(key, value)
	if err != nil {
		ServerError(w, err)
		return
	}
	if !ok {
		w.WriteHeader(http.StatusCreated)
	}
	w.Write(body)
}

func Key(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s request to %s", r.Method, r.URL.String())
	VersionHeader(w)
	switch r.Method {
	case "GET":
		GetKey(w, r)
	case "POST", "PUT":
		SetKey(w, r)
	case "DELETE":
		DelKey(w, r)
	case "HEAD":
		w.Header().Add("content-length", "0")
		w.WriteHeader(http.StatusOK)
		return
	default:
		log.Print("received unsupported request for method ",
			r.Method)
		NotImplemented(w, r)

	}
}

func main() {
	http.HandleFunc("/data", Key)
	http.HandleFunc("/data/", Key)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
