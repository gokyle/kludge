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

func NotImplemented(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	msg := "Method " + r.Method + " not implemented."
	w.Write([]byte(msg))
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
	key := KeyID(r)
	//	if key == "" {
	//		ListKeys(w, r)
	//		return
	//	}
	body, err := getKey(key)
	if err != nil {
		ServerError(w, err)
		return
	}
	w.Write(body)
}

func DelKey(w http.ResponseWriter, r *http.Request) {
	body, err := delKey(KeyID(r))
	if err != nil {
		ServerError(w, err)
		return
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
	body, err := setKey(key, value)
	if err != nil {
		ServerError(w, err)
		return
	}
	w.Write(body)
}

func Key(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s request to %s", r.Method, r.URL.String())
	if r.URL.Path == "/key" || r.URL.Path == "/key/" {
		ListKeys(w, r)
	} else {
		switch r.Method {
		case "GET":
			GetKey(w, r)
		case "POST", "PUT":
			SetKey(w, r)
		case "DELETE":
			DelKey(w, r)
		default:
			NotImplemented(w, r)

		}
	}
}

func main() {
	http.HandleFunc("/key", Key)
	http.HandleFunc("/key/", Key)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
