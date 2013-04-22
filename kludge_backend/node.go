package main

import (
	"flag"
	"github.com/gokyle/goconfig"
	"github.com/jmhodges/levigo"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var (
	dataStore  string // the filepath kludge should store data in
	dbOpts     *levigo.Options
	ldb        *levigo.DB
	listenAddr = ":5987"
	poolSize   = 4
	reqBuf     = 16
)

func init() {
	configFile := flag.String("f", "etc/kludge/backendrc",
		"path to configuration file")
	flag.Parse()

	var cfg map[string]string
	if cfgmap, err := goconfig.ParseFile(*configFile); err != nil {
		log.Fatal(err.Error())
	} else {
		cfg = cfgmap["default"]
	}

	dataStore = cfg["datastore"]
	if dataStore == "" {
		log.Fatal("no datastore specified")
	}

	if cfgAddr, ok := cfg["listen"]; ok {
		listenAddr = cfgAddr
	}

	if cfgReqBuf, ok := cfg["request_buffer"]; ok {
		var err error

		reqBuf, err = strconv.Atoi(cfgReqBuf)
		if err != nil {
			log.Printf("invalid value %s for request buffer: %s",
				cfgReqBuf, err.Error())
		}
	}

	if cfgPSize, ok := cfg["pool_size"]; ok {
		var err error

		poolSize, err = strconv.Atoi(cfgPSize)
		if err != nil {
			log.Printf("invalid value %s for pool size: %s",
				cfgPSize, err.Error())
		}
	}
	dbOpts = levigo.NewOptions()
	dbOpts.SetCache(levigo.NewLRUCache(3 << 20))
	dbOpts.SetCreateIfMissing(true)
}

func main() {
	var err error
	sigc := make(chan os.Signal, 1)

	log.Println("starting kludge")
	ldb, err = levigo.Open(dataStore, dbOpts)
	if err != nil {
		log.Fatal("Failed to start kludge backend: ", err.Error())
	}
	defer ldb.Close()

	go startPool()
	go listener()
	signal.Notify(sigc, os.Kill, os.Interrupt, syscall.SIGTERM)
	<-sigc

	// the worker pool is managed in pool.go.
	if reqQ != nil {
		close(reqQ)
		log.Println("giving workers time to complete")
		<-time.After(250 * time.Millisecond)
	}
	log.Println("kludge is shutting down")
}
