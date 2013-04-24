// Package logsrvc is a client library for the kludge remote logging server.
package logsrvc

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

// Logger is value capable of sending logs to the log server.
type Logger struct {
	conn *net.TCPConn
	node string
}

// Connect is used to establish a connection to a logging server.
func Connect(node, addr string) (logger *Logger, err error) {
	logger = new(Logger)

	if addr != "" {
		var logSrvAddr *net.TCPAddr
		logSrvAddr, err = net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			log.Printf("failed to resolve address %s: %s", addr,
				err.Error())
			return
		}

		logger.node = node
		logger.conn, err = net.DialTCP("tcp", nil, logSrvAddr)
		if err != nil {
			log.Printf("couldn't connect to %s: %s", addr, err.Error())
			return
		}

		logger.conn.SetKeepAlive(true)
		log.SetOutput(logger.conn)
	}
	log.SetFlags(0)
	return
}

// Print, Println, Printf, Fatalf, Fatal, and Fatalln all wrap the
// corresponding log functions. They also print the message to the
// console; it is done this way because Upstart adds its own timestamp
// to log messages, making additional timestamps unnecessary and
// causing them to clutter up the logs.
func (logger *Logger) Print(args ...interface{}) {
	ts := time.Now().Unix()
	precur := fmt.Sprintf("%s %d ", logger.node, ts)

	logargs := make([]interface{}, 0)
	logargs = append(logargs, precur)
	logargs = append(logargs, args...)

	fmt.Println(logargs...)

	if logger.conn != nil {
		log.Print(logargs...)
	}
}

func (logger *Logger) Printf(format string, args ...interface{}) {
	ts := time.Now().Unix()
	precur := fmt.Sprintf("%s %d ", logger.node, ts)

	if logger.conn != nil {
		log.Printf(precur+format, args...)
	}
	fmt.Printf(precur+format+"\n", args...)
}

func (logger *Logger) Println(args ...interface{}) {
	ts := time.Now().Unix()
	precur := fmt.Sprintf("%s %d", logger.node, ts)

	logargs := make([]interface{}, 0)
	logargs = append(logargs, precur)
	logargs = append(logargs, args...)

	fmt.Println(logargs...)
	if logger.conn != nil {
		log.Println(logargs...)
	}
}

func (logger *Logger) Fatal(args ...interface{}) {
	logger.Print(args...)
	os.Exit(1)
}

func (logger *Logger) Fatalf(format string, args ...interface{}) {
	logger.Printf(format, args...)
	os.Exit(1)
}

func (logger *Logger) Fatalln(args ...interface{}) {
	logger.Println(args...)
	os.Exit(1)
}

// Shutdown ensures a clean shutdown.
func (logger *Logger) Shutdown() {
	logger.conn.Close()
}
