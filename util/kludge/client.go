package main

import (
	"flag"
	"fmt"
	"github.com/gokyle/kludge/kludge_client"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	addr := flag.String("a", "127.0.0.1:8080", "Kludge API server address")
	flFile := flag.String("f", "", "a file to use for a SET operation")
	flKey := flag.String("k", "", "the key to act on")
	flStr := flag.Bool("s", false, "show value as string")
	flVal := flag.String("v", "", "the value for a SET operation")
	flOp := flag.String("x", "GET", "operation (GET, SET, DEL, LST)")
	flag.Parse()

	*flOp = strings.ToUpper(*flOp)

	switch *flOp {
	case "GET":
		get(*addr, *flKey, *flStr)
	case "SET":
		var value []byte
		if *flFile != "" {
			var err error
			value, err = ioutil.ReadFile(*flFile)
			if err != nil {
				fmt.Println("[!]", err.Error())
				os.Exit(1)
			}
		} else {
			value = []byte(*flVal)
		}
		set(*addr, *flKey, value, *flStr)
	case "DEL":
		del(*addr, *flKey, *flStr)
	case "LST":
		lst(*addr)
	default:
		fmt.Printf("[!] %s is not a supported operation.\n", *flOp)
		fmt.Println("\tsupported operations: GET, SET, DEL, LST")
		os.Exit(1)
	}
}

func get(addr, key string, toString bool) {
	if key == "" {
		fmt.Println("[!] no key chosen. GET requires a key.")
		os.Exit(1)
	}
	ds, err := kludge.Connect(addr, nil)
	if err != nil {
		fmt.Println("[!] error connecting to datastore:", err.Error())
		os.Exit(1)
	}
	value, ok, err := ds.Get(key)
	if err != nil {
		fmt.Println("[!] error getting value:", err.Error())
		os.Exit(1)
	}

	fmt.Printf("[ %s ]\n", key)
	if ok {
		fmt.Printf("\tpresent in datastore\n")
	} else {
		fmt.Printf("\tkey not in datastore\n")
		return
	}
	fmt.Printf("\tvalue: ")
	if toString {
		fmt.Println(string(value))
	} else {
		fmt.Printf("%v\n", value)
	}
}

func set(addr, key string, value []byte, toString bool) {
	if key == "" {
		fmt.Println("[!] no key chosen. SET requires a key.")
		os.Exit(1)
	}
	ds, err := kludge.Connect(addr, nil)
	if err != nil {
		fmt.Println("[!] error connecting to datastore:", err.Error())
		os.Exit(1)
	}
	prev, ok, err := ds.Set(key, value)
	fmt.Printf("[ %s ]\n", key)
	if ok {
		fmt.Printf("\tpresent in datastore\n")
		fmt.Printf("\tprev value: ")
		if toString {
			fmt.Println(string(prev))
		} else {
			fmt.Printf("%v\n", prev)
		}
	} else {
		fmt.Printf("\tkey not in datastore\n")
	}
	fmt.Println("\tkey set successfully.")
}

func del(addr, key string, toString bool) {
	if key == "" {
		fmt.Println("[!] no key chosen. DEL requires a key.")
		os.Exit(1)
	}
	ds, err := kludge.Connect(addr, nil)
	if err != nil {
		fmt.Println("[!] error connecting to datastore:", err.Error())
		os.Exit(1)
	}
	prev, ok, err := ds.Del(key)
	fmt.Printf("[ %s ]\n", key)
	if ok {
		fmt.Printf("\tpresent in datastore\n")
		fmt.Printf("\tprev value: ")
		if toString {
			fmt.Println(string(prev))
		} else {
			fmt.Printf("%v\n", prev)
		}
	        fmt.Println("\tkey deleted successfully.")
	} else {
		fmt.Printf("\tkey not in datastore\n")
	}
}

func lst(addr string) {
	ds, err := kludge.Connect(addr, nil)
	if err != nil {
		fmt.Println("[!] error connecting to datastore:", err.Error())
		os.Exit(1)
	}
        keys, err := ds.List()
        if err != nil {
                fmt.Printf("[!] failed to get keys:", err.Error())
                os.Exit(1)
        }
        keyList := strings.Join(keys, ", ")
        fmt.Printf("keys in datastore: %s\n", keyList)
}
