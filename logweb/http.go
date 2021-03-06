package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/gokyle/goconfig"
	"github.com/gokyle/kludge/logsrv/logsrvc"
	"github.com/gokyle/uuid"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var (
	addr     string
	dbFile   string
	nodeID   string
	logger   *logsrvc.Logger
	tplDir   string
	assetDir string
)

func initLogging(cfgmap goconfig.ConfigMap) (regen bool) {
	var err error
	cfg, ok := cfgmap["logging"]
	if !ok {
		fmt.Println("no logging information!")
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

	logger, err = logsrvc.Connect("logweb:"+nodeID, logserver)
	if err != nil {
		fmt.Println("failed to set up log host:", err.Error())
		os.Exit(1)
	}
	return
}

func initServer(cfgmap goconfig.ConfigMap) (regen bool) {
	var ok bool

	cfg, ok := cfgmap["server"]
	if !ok {
		logger.Fatal("bad config file: missing section 'server'")
	}

	if addr, ok = cfg["address"]; !ok {
		addr = ":8080"
	}

	if dbFile, ok = cfg["database"]; !ok {
		logger.Fatal("no database file configured")
	}

	if assetDir, ok = cfg["assets"]; !ok {
		assetDir = "assets"
	}

	if tplDir, ok = cfg["templates"]; !ok {
		tplDir = "templates"
	}
	return false
}

func init() {
	configFile := flag.String("f", "logwebrc",
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

	if initServer(cfg) {
		updateConfig(cfg, *configFile)
	}
}

func templatePath(name string) string {
	return filepath.Join(tplDir, name)
}

func updateConfig(cfg goconfig.ConfigMap, cfgFile string) {
	err := cfg.WriteFile(cfgFile)
	if err != nil {
		fmt.Println("failed to update config file:",
			err.Error())
		os.Exit(1)
	}
}

func serverError(w http.ResponseWriter, err error) {
	logger.Println("error handling request:", err.Error())
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}

func response(w http.ResponseWriter, r *http.Request) {
	logger.Printf("%s request from %s to %s", r.Method,
		r.RemoteAddr, r.URL.Path)

	var nodes []*responseTime
	nodelst, err := responseNodes()
	if err != nil {
		serverError(w, err)
		return
	}

	for _, nodeName := range nodelst {
		node, err := getNodeAverages(nodeName)
		if err != nil {
			serverError(w, err)
			return
		}
		nodes = append(nodes, node)
	}

	page, err := ioutil.ReadFile(templatePath("response.html"))
	if err != nil {
		serverError(w, err)
		return
	}
	respTpl := template.New("response")
	respTpl, err = respTpl.Parse(string(page))
	if err != nil {
		serverError(w, err)
		return
	}

	buf := new(bytes.Buffer)
	err = respTpl.Execute(w, nodes)
	if err != nil {
		serverError(w, err)
		return
	}
	w.Write(buf.Bytes())
}

func logs_all(w http.ResponseWriter, r *http.Request) {
	logger.Printf("%s request from %s to %s", r.Method,
		r.RemoteAddr, r.URL.Path)
	start := time.Now().UnixNano()

	entries, err := logsFrom(0)
	if err != nil {
		serverError(w, err)
		return
	}

	if filepath.Ext(r.URL.Path) == ".txt" {
		dump_log(entries, w)
		rtime := (time.Now().UnixNano() - start) / 1000.0
		logger.Printf("GET response time: %dus", rtime)
		return
	}

	page, err := ioutil.ReadFile(templatePath("logs.html"))
	if err != nil {
		serverError(w, err)
		return
	}

	tpl := template.New("logs")
	tpl, err = tpl.Parse(string(page))
	if err != nil {
		serverError(w, err)
		return
	}

	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, entries)
	if err != nil {
		serverError(w, err)
		return
	}
	w.Write(buf.Bytes())
	rtime := (time.Now().UnixNano() - start) / 1000.0
	logger.Printf("GET response time: %dus", rtime)
}

func dump_log(entries []*logEntry, w http.ResponseWriter) {
	logger.Printf("dumping %d entries to text file", len(entries))
	w.Header().Add("content-type", "text-plain")
	for i := len(entries) - 1; i != 0; i-- {
		logLine := fmt.Sprintf("%s %s %s\n", entries[i].Time,
			entries[i].Node, entries[i].Msg)
		w.Write([]byte(logLine))
	}
}

func logs_hourly(w http.ResponseWriter, r *http.Request) {
	logger.Printf("%s request from %s to %s", r.Method,
		r.RemoteAddr, r.URL.Path)

	hour, err := time.ParseDuration("-1h")
	if err != nil {
		serverError(w, err)
		return
	}
	when := time.Now().Add(hour).Unix()

	entries, err := logsFrom(when)
	if err != nil {
		serverError(w, err)
		return
	}

	if filepath.Ext(r.URL.Path) == ".txt" {
		dump_log(entries, w)
		return
	}

	page, err := ioutil.ReadFile(templatePath("logs.html"))
	if err != nil {
		serverError(w, err)
		return
	}

	tpl := template.New("logs")
	tpl, err = tpl.Parse(string(page))
	if err != nil {
		serverError(w, err)
		return
	}

	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, entries)
	if err != nil {
		serverError(w, err)
		return
	}
	w.Write(buf.Bytes())
}

func root(w http.ResponseWriter, r *http.Request) {
	logger.Printf("%s request from %s to %s", r.Method,
		r.RemoteAddr, r.URL.Path)

	indexFile, err := ioutil.ReadFile(templatePath("index.html"))
	if err != nil {
		serverError(w, err)
		return
	}

	respTpl := template.New("index")
	respTpl, err = respTpl.Parse(string(indexFile))
	if err != nil {
		serverError(w, err)
		return
	}

	buf := new(bytes.Buffer)
	err = respTpl.Execute(buf, nil)
	if err != nil {
		logger.Printf("buf: %s", string(buf.Bytes()))
		serverError(w, err)
		return
	}
	w.Write(buf.Bytes())
}

func main() {
	http.HandleFunc("/response/", response)
	http.HandleFunc("/logs/all", logs_all)
	http.HandleFunc("/logs/all.txt", logs_all)
	http.HandleFunc("/logs/hourly", logs_hourly)
	http.HandleFunc("/logs/hourly.txt", logs_hourly)
	http.HandleFunc("/", root)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetDir))))
	logger.Printf("logweb listing on http://%s/", addr)
	go http.ListenAndServe(addr, nil)
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Kill, os.Interrupt, syscall.SIGTERM)
	fmt.Println("[+] waiting for shutdown signal")
	<-sigc
	logger.Println("logweb shutdown")
}
