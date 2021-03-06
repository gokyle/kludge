package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"net"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type logEntry struct {
	Node string
	Time int64
	Msg  string
}

type clientEntry struct {
	Addr   string
	Time   int64
	Online int
}

var (
	tables   map[string]string
	tsFormat = "2006/01/02 15:04:05"
	addr     string
	dbFile   string
	logChan  chan *logEntry
	cliChan  chan *clientEntry
	rawChan  chan string
)

var logSplit = regexp.MustCompile("^([\\w-:]+) (\\d+) (.+)$")
var responseCheck = regexp.MustCompile("(\\w+) response time: (\\d+)us$")

func init() {
	flLogBuffer := flag.Uint("b", 16, "log entries to buffer")
	flDbFile := flag.String("f", "logs.db", "database file")
	port := flag.Uint("p", 5988, "port to listen on")
	flag.Parse()

	addr = fmt.Sprintf(":%d", *port)
	dbFile = *flDbFile
	logChan = make(chan *logEntry, *flLogBuffer)
	cliChan = make(chan *clientEntry, *flLogBuffer)
	rawChan = make(chan string, *flLogBuffer)

	tables = make(map[string]string, 0)
	tables["entries"] = `CREATE TABLE entries (id integer primary key,
                node text,
                timestamp integer,
                message string)`
	tables["response_time"] = `CREATE TABLE response_time (
                id integer primary key,
                node text,
                timestamp integer,
                microsec integer,
                operation text)`
	tables["clients"] = `CREATE TABLE clients (
                id integer primary key,
                address text,
                timestamp integer,
                online integer)`
	tables["raw_logs"] = `CREATE TABLE raw_logs (
                id integer primary key,
                timestamp integer,
                entry text)`
}

func main() {
	dbSetup()
	go listen()
	go log()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Kill, os.Interrupt, syscall.SIGTERM)
	fmt.Println("[+] waiting for shutdown signal")
	<-sigc
	fmt.Println("[+] closing log channel")
	close(logChan)
	fmt.Println("[+] closing log channel")
	close(cliChan)
	<-time.After(100 * time.Millisecond)

	fmt.Println("[+] logsrv shutting down")
	os.Exit(0)
}

func listen() {
	fmt.Println("[+] start TCP server")
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[!] failed to resolve TCP address:", err.Error())
		os.Exit(1)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[!] failed to set up TCP listener:", err.Error())

	} else {
		defer listener.Close()
		fmt.Println("[+] listening for clients")
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("[!] TCP error:", err.Error())
				continue
			}
			go processMessage(conn)
		}
	}
	fmt.Println("[+] TCP server shuts down")
}

func processMessage(conn net.Conn) {
	client := new(clientEntry)
	client.Addr = conn.RemoteAddr().String()
	client.Time = time.Now().UTC().Unix()
	client.Online = 1
	cliChan <- client

	fmt.Println("[+] client connected:", conn.RemoteAddr())
	defer conn.Close()
	r := bufio.NewReader(conn)

	for {
		msg, err := r.ReadString(0x0a)
		if err != nil {
			if err != io.EOF {
				fmt.Println("[!] error reading from client:",
					err.Error())
			}
			break
		} else if msg == "" {
			continue
		}
		msg = strings.Trim(string(msg), "\n \t")
		fmt.Println("-- ", msg)
		rawChan <- msg

		nodeID := logSplit.ReplaceAllString(msg, "$1")
		dateString := logSplit.ReplaceAllString(msg, "$2")
		logMsg := logSplit.ReplaceAllString(msg, "$3")
		timestamp, err := strconv.ParseInt(dateString, 10, 64)
		if err != nil {
			fmt.Println("[!] error parsing timestamp:", err.Error())
			continue
		}

		le := &logEntry{nodeID, timestamp, logMsg}
		logChan <- le
	}
	fmt.Println("[+] client disconnected:", conn.RemoteAddr())
	client.Online = 0
	cliChan <- client
}

func log() {
	fmt.Println("[+] start log listener")
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		fmt.Println("[!] failed to open DB file:", err.Error())
		os.Exit(1)
	}
	defer db.Close()

	for {
		select {
		case le, ok := <-logChan:
			if !ok {
				return
			}
			writeLogEntry(db, le)
		case client, ok := <-cliChan:
			if !ok {
				return
			}
			writeClientEntry(db, client)
		case message, ok := <-rawChan:
			if !ok {
				return
			}
			writeRawEntry(db, message)
		}
	}
}

func writeLogEntry(db *sql.DB, le *logEntry) {
	_, err := db.Exec(`insert into entries (node, timestamp, message)
            values (?, ?, ?)`, le.Node, le.Time, le.Msg)
	if err != nil {
		fmt.Println("[!] database error:", err.Error())
		return
	}

	if responseCheck.MatchString(le.Msg) {
		opName := responseCheck.ReplaceAllString(le.Msg, "$1")
		respString := responseCheck.ReplaceAllString(le.Msg, "$2")
		rTime, err := strconv.Atoi(respString)
		if err != nil {
			fmt.Println("[!] error reading response time:", err.Error())
			return
		}
		_, err = db.Exec(`insert into response_time
                        (node, timestamp, microsec, operation)
                        values (?, ?, ?, ?)`, le.Node, le.Time, rTime, opName)
		if err != nil {
			fmt.Println("[!] error writing to database:", err.Error())
		}
	}
}

func writeClientEntry(db *sql.DB, cli *clientEntry) {
	_, err := db.Exec(`insert into clients (address, timestamp, online)
                values (?, ?, ?)`, cli.Addr, cli.Time, cli.Online)
	if err != nil {
		fmt.Println("[!] database error:", err.Error())
	}
}

func writeRawEntry(db *sql.DB, msg string) {
	_, err := db.Exec(`insert into raw_logs (timestamp, entry)
                values (?, ?)`, time.Now().Unix(), msg)
	if err != nil {
		fmt.Println("[!] database error:", err.Error())
	}
}

func dbSetup() {
	fmt.Println("[+] checking tables")
	for tableName, tableSQL := range tables {
		fmt.Printf("\t[*] table %s\n", tableName)
		checkTable(tableName, tableSQL)
	}
	fmt.Println("[+] finished checking database")
}

func checkTable(tableName, tableSQL string) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		fmt.Println("[!] failed to open DB file:", err.Error())
		os.Exit(1)
	}
	defer db.Close()

	rows, err := db.Query(`select sql from sqlite_master
                               where type='table' and name=?`, tableName)
	if err != nil {
		fmt.Println("[!] error looking up table:", err.Error())
		os.Exit(1)
	}

	var tblSql string
	for rows.Next() {
		err = rows.Scan(&tblSql)
		break
	}
	rows.Close()

	if err != nil {
		fmt.Println("[!] error reading database:", err.Error())
		os.Exit(1)
	} else if tblSql == "" {
		fmt.Println("\t\t[+] creating table")
		_, err = db.Exec(tableSQL)
		if err != nil {
			fmt.Println("[!] error creating table:", err.Error())
			os.Exit(1)
		}
	} else if tblSql != tableSQL {
		fmt.Println("\t\t[+] schema out of sync")
		_, err = db.Exec("drop table " + tableName)
		if err != nil {
			fmt.Println("[!] error dropping table:", err.Error())
			os.Exit(1)
		}
		_, err = db.Exec(tableSQL)
		if err != nil {
			fmt.Println("[!] error creating table:", err.Error())
			os.Exit(1)
		}
		fmt.Printf("\t[+] table %s updated\n", tableName)
	}
}
