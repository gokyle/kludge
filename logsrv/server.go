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

var (
	tsFormat = "2006/01/02 15:04:05"
	addr     string
	dbFile   string
	logChan  chan *logEntry
)

var logSplit = regexp.MustCompile("^([\\w-:]+) (\\d{4}/\\d{2}/\\d{2} \\d{2}:\\d{2}:\\d{2}) (.+)$")
var responseCheck = regexp.MustCompile("response time: (\\d+)us$")

func init() {
	flLogBuffer := flag.Uint("b", 16, "log entries to buffer")
	flDbFile := flag.String("f", "logs.db", "database file")
	port := flag.Uint("p", 5988, "port to listen on")
	flag.Parse()

	addr = fmt.Sprintf(":%d", *port)
	dbFile = *flDbFile
	logChan = make(chan *logEntry, *flLogBuffer)
}

func main() {
	dbSetup()
	go log()
	go listen()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Kill, os.Interrupt, syscall.SIGTERM)
	<-sigc
	close(logChan)
	<-time.After(100 * time.Millisecond)

	os.Exit(1)
}

func listen() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[!] failed to resolve TCP address:", err.Error())
		os.Exit(1)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[!] failed to set up TCP listener:", err.Error())
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("[!] TCP error:", err.Error())
			continue
		}
		go processMessage(conn)
	}

}

func processMessage(conn net.Conn) {
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
			return
		} else if msg == "" {
			return
		}
		msg = strings.Trim(string(msg), "\n \t")
		fmt.Println("-- ", msg)

		nodeID := logSplit.ReplaceAllString(msg, "$1")
		dateString := logSplit.ReplaceAllString(msg, "$2")
		logMsg := logSplit.ReplaceAllString(msg, "$3")
		tm, err := time.Parse(tsFormat, dateString)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[!] error parsing time %s: %s\n",
				dateString, err.Error())
			return
		}
		le := &logEntry{nodeID, tm.UTC().Unix(), logMsg}
		logChan <- le
	}
}

func log() {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		fmt.Println("[!] failed to open DB file:", err.Error())
		os.Exit(1)
	}
	defer db.Close()

	for {
		le, ok := <-logChan
		if !ok {
			fmt.Println("[+] shutting down database listener")
			return
		}
		_, err := db.Exec("insert into entries values (?, ?, ?)",
			le.Node, le.Time, le.Msg)
		if err != nil {
			fmt.Println("[!] database error:", err.Error())
			continue
		}

		if responseCheck.MatchString(le.Msg) {
			respString := responseCheck.ReplaceAllString(le.Msg, "$1")
			rTime, err := strconv.Atoi(respString)
			if err != nil {
				fmt.Println("[!] error reading response time:", err.Error())
				return
			}
			_, err = db.Exec("insert into response_times values (?, ?, ?)",
				le.Node, le.Time, rTime)
			if err != nil {
				fmt.Println("[!] error writing to database:", err.Error())
			}
		}

	}
}

func logResponseTime(db *sql.DB, le *logEntry) {
}

func dbSetup() {
	entryTable()
	respTable()
}

func entryTable() {
	const createSql = `CREATE TABLE entries (node text, timestamp integer, message string)`
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		fmt.Println("[!] failed to open DB file:", err.Error())
		os.Exit(1)
	}
	defer db.Close()

	rows, err := db.Query(`select sql from sqlite_master
                               where type='table' and name='entries'`)
	if err != nil {
		fmt.Println("[!] error reading database:", err.Error())
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
		fmt.Println("[+] creating table")
		_, err = db.Exec(createSql)
		if err != nil {
			fmt.Println("[!] error creating table:", err.Error())
			os.Exit(1)
		}
	} else if tblSql != createSql {
		fmt.Println("[+] schema out of sync")
		_, err = db.Exec(`drop table entries`)
		if err != nil {
			fmt.Println("[!] error dropping table:", err.Error())
			os.Exit(1)
		}
		_, err = db.Exec(createSql)
		if err != nil {
			fmt.Println("[!] error creating table:", err.Error())
			os.Exit(1)
		}
	}
}

func respTable() {
	const createSql = `CREATE TABLE response_times (node text, timestamp integer, microsec integer)`
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		fmt.Println("[!] failed to open DB file:", err.Error())
		os.Exit(1)
	}
	defer db.Close()

	rows, err := db.Query(`select sql from sqlite_master
                               where type='table' and name='response_times'`)
	if err != nil {
		fmt.Println("[!] error reading database:", err.Error())
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
		fmt.Println("[+] creating table")
		_, err = db.Exec(createSql)
		if err != nil {
			fmt.Println("[!] error creating table:", err.Error())
			os.Exit(1)
		}
	} else if tblSql != createSql {
		fmt.Println("[+] schema out of sync")
		_, err = db.Exec(`drop table response_times`)
		if err != nil {
			fmt.Println("[!] error dropping table:", err.Error())
			os.Exit(1)
		}
		_, err = db.Exec(createSql)
		if err != nil {
			fmt.Println("[!] error creating table:", err.Error())
			os.Exit(1)
		}
	}
}
