package main

import (
        "log"
        "github.com/jmhodges/levigo"
        "os"
        "os/signal"
        "syscall"
)

var (
        dataStore string // the filepath kludge should store data in
        dbOpts *levigo.Options
        ldb *levigo.DB
)

func init() {
        dataStore = "data"
        dbOpts = levigo.NewOptions()
        dbOpts.SetCache(levigo.NewLRUCache(3<<30))
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

        signal.Notify(sigc, os.Kill, os.Interrupt, syscall.SIGTERM)
        <-sigc

        log.Println("kludge is shutting down")
}
