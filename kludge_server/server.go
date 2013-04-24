package main

import (
	"flag"
	"fmt"
	"github.com/gokyle/goconfig"
	"github.com/gokyle/kludge/common"
	"github.com/gokyle/kludge/logsrv/logsrvc"
	"github.com/gokyle/uuid"
	"io"
	"net/http"
	"os"
	"regexp"
)

var (
	configFile string
	address    string
	nodeID     string
	logserver  string
	logger     *logsrvc.Logger
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
		logger.Printf("request for %s failed: %s", r.URL.String(),
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
	logger.Printf("%s request to %s", r.Method, r.URL.String())
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
		logger.Print("received unsupported request for method ",
			r.Method)
		NotImplemented(w, r)

	}
}

func main() {
	defer logger.Shutdown()
	address = "127.0.0.1:8080"
	http.HandleFunc("/data", Key)
	http.HandleFunc("/data/", Key)
	logger.Println("serving on", address)
	logger.Fatal(http.ListenAndServe(address, nil))
}

func init() {
	cfgFile := flag.String("f", "/etc/kludge/serverrc",
		"configuration file")
	flag.Parse()

	cfg, err := goconfig.ParseFile(*cfgFile)
	if err != nil {
		fmt.Println("failed to parse configuration file:", err.Error())
		os.Exit(1)
	}
	regen := initLogging(cfg["logging"])
	if regen {
		cfg["logging"]["node_id"] = nodeID
		err = cfg.WriteFile(*cfgFile)
		if err != nil {
			fmt.Println("failed to update config file:",
				err.Error())
			os.Exit(1)
		}
	}
}

func initLogging(cfg map[string]string) (regen bool) {
	var err error

	nodeID = cfg["node_id"]
	if nodeID == "" {
		nodeID, err = uuid.GenerateV4String()
		if err != nil {
			fmt.Println("failed to generate node ID:", err.Error())
			os.Exit(1)
		}
		regen = true
	}
	logserver = cfg["loghost"]

	logger, err = logsrvc.Connect("srv:"+nodeID, logserver)
	if err != nil {
		fmt.Println("failed to set up log host:", err.Error())
		os.Exit(1)
	}
	return
}
