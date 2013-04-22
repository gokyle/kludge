package kludge

import (
	"fmt"
	"os"
	"testing"
)

var TestServer = "127.0.0.1:8080"

func init() {
	db, err := Connect(TestServer, nil)
	if err != nil {
		fmt.Println("[!] database couldn't be reached")
		os.Exit(1)
	}
	db.Set("foo", []byte("bar"))
	db.Del("baz")
}

func TestConnect(t *testing.T) {
	_, err := Connect(TestServer, nil)
	if err != nil {
		fmt.Println("[!] connect failed: ", err.Error())
		t.FailNow()
	}
}

func TestGet(t *testing.T) {
	db, err := Connect(TestServer, nil)
	if err != nil {
		fmt.Println("[!] connect failed: ", err.Error())
		t.FailNow()
	}

	value, ok, err := db.Get("foo")
	if err != nil {
		fmt.Println("GET failed: ", err.Error())
		t.FailNow()
	} else if !ok {
		fmt.Println("expected key to be present!")
		t.FailNow()
	} else if string(value) != "bar" {
		fmt.Println("unexpected value found")
		t.FailNow()
	}
}

func TestSet(t *testing.T) {
	db, err := Connect(TestServer, nil)
	if err != nil {
		fmt.Println("[!] connect failed: ", err.Error())
		t.FailNow()
	}

	value, ok, err := db.Set("bar", []byte("baz"))
	if err != nil {
		fmt.Println("SET failed: ", err.Error())
		t.FailNow()
	} else if ok {
		fmt.Println("expected key to not be present!")
		t.FailNow()
	} else if string(value) != "" {
		fmt.Println("unexpected value found")
		t.FailNow()
	}
}

func TestDel(t *testing.T) {
	db, err := Connect(TestServer, nil)
	if err != nil {
		fmt.Println("[!] connect failed: ", err.Error())
		t.FailNow()
	}

	value, ok, err := db.Del("bar")
	if err != nil {
		fmt.Println("DEL failed: ", err.Error())
		t.FailNow()
	} else if !ok {
		fmt.Println("expected key to be present!")
		t.FailNow()
	} else if string(value) != "baz" {
		fmt.Println("unexpected value found")
		t.FailNow()
	}
}
