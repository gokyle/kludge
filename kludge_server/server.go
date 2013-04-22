package main

import (
	"log"
	"net/http"
	"regexp"
)

var keyIDRegexp = regexp.MustCompile("^/key/(.+)$")

func ServerError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}

func GetKey(w http.ResponseWriter, r *http.Request) {
	keyID := keyIDRegexp.ReplaceAllString(r.URL.Path, "$1")
	body, err := getKey(keyID)
	if err != nil {
		ServerError(w, err)
		return
	}
	w.Write(body)
}

func main() {
	http.HandleFunc("/key/", GetKey)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
