package main

import (
	"flag"
	"fmt"
	"github.com/gokyle/goconfig"
	"github.com/gokyle/kludge/logsrv/logsrvc"
	"github.com/gokyle/uuid"
	"github.com/jmhodges/levigo"
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
	logger     *logsrvc.Logger
	nodeID     string
	listenAddr = ":5987"
	poolSize   = 4
	reqBuf     = 16
)

func initLogging(cfgmap goconfig.ConfigMap) (regen bool) {
	var err error
	cfg, ok := cfgmap["logging"]
	if !ok {
		fmt.Println("[!] no logging information!")
		os.Exit(1)
	}

	nodeID = cfg["node_id"]
	if nodeID == "" {
		nodeID, err = uuid.GenerateV4String()
		if err != nil {
			fmt.Println("failed to generate node ID:", err.Error())
			os.Exit(1)
		}
		regen = true
		cfgmap["logging"]["node_id"] = nodeID
	}
	logserver := cfg["loghost"]

	logger, err = logsrvc.Connect("node:"+nodeID, logserver)
	if err != nil {
		fmt.Println("failed to set up log host:", err.Error())
		os.Exit(1)
	}
	return
}

func initDatastore(cfgmap goconfig.ConfigMap) (regen bool) {
	cfg, ok := cfgmap["datastore"]
	if !ok {
		fmt.Println("[!] no datastore information!")
		os.Exit(1)
	}

	dataStore = cfg["datastore"]
	if dataStore == "" {
		logger.Fatal("no datastore specified")
	}

	if cfgAddr, ok := cfg["listen"]; ok {
		listenAddr = cfgAddr
	}

	if cfgReqBuf, ok := cfg["request_buffer"]; ok {
		var err error

		reqBuf, err = strconv.Atoi(cfgReqBuf)
		if err != nil {
			logger.Printf("invalid value %s for request buffer: %s",
				cfgReqBuf, err.Error())
		}
	}

	if cfgPSize, ok := cfg["pool_size"]; ok {
		var err error

		poolSize, err = strconv.Atoi(cfgPSize)
		if err != nil {
			logger.Printf("invalid value %s for pool size: %s",
				cfgPSize, err.Error())
		}
	}
	dbOpts = levigo.NewOptions()
	dbOpts.SetCache(levigo.NewLRUCache(3 << 20))
	dbOpts.SetCreateIfMissing(true)
	return false
}

func updateConfig(cfg goconfig.ConfigMap, cfgFile string) {
	err := cfg.WriteFile(cfgFile)
	if err != nil {
		fmt.Println("failed to update config file:",
			err.Error())
		os.Exit(1)
	}

}

func init() {
	configFile := flag.String("f", "etc/kludge/backendrc",
		"path to configuration file")
	flag.Parse()

	cfg, err := goconfig.ParseFile(*configFile)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if initLogging(cfg) {
		updateConfig(cfg, *configFile)
	}

	if initDatastore(cfg) {
		updateConfig(cfg, *configFile)
	}
}

func main() {
	var err error
	sigc := make(chan os.Signal, 1)

	logger.Println("starting kludge")
	ldb, err = levigo.Open(dataStore, dbOpts)
	if err != nil {
		logger.Fatal("Failed to start kludge backend: ", err.Error())
	}
	defer ldb.Close()

	go startPool()
	go listener()
	signal.Notify(sigc, os.Kill, os.Interrupt, syscall.SIGTERM)
	<-sigc

	// the worker pool is managed in pool.go.
	if reqQ != nil {
		close(reqQ)
		logger.Println("giving workers time to complete")
		<-time.After(250 * time.Millisecond)
	}
	logger.Println("kludge is shutting down")
}
