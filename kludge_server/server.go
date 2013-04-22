package main

import (
	"io"
	"log"
	"net/http"
	"regexp"
)

var keyIDRegexp = regexp.MustCompile("^/key/(.+)$")

func ServerError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}

func KeyID(r *http.Request) string {
	return keyIDRegexp.ReplaceAllString(r.URL.Path, "$1")
}

func GetKey(w http.ResponseWriter, r *http.Request) {
	body, err := getKey(KeyID(r))
	if err != nil {
		ServerError(w, err)
		return
	}
	w.Write(body)
}

func SetKey(w http.ResponseWriter, r *http.Request) {
	key := KeyID(r)
	defer r.Body.Close()

	// TODO: check CL > 0
	value := make([]byte, r.ContentLength)
	_, err := io.ReadFull(r.Body, value)
	if err != nil {
		log.Printf("request for %s failed: %s", r.URL.String(),
			err.Error())
		ServerError(w, err)
		return
	}
	body, err := setKey(key, value)
	if err != nil {
		ServerError(w, err)
		return
	}
	w.Write(body)
}

func Key(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/key" {
	} else {
		switch r.Method {
		case "GET":
			GetKey(w, r)
		case "POST", "PUT":
			SetKey(w, r)
		}
	}
}

func main() {
	http.HandleFunc("/key/", GetKey)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
