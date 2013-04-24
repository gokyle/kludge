// Package logsrvc is a client library for the kludge remote logging server.
package logsrvc

import (
	"fmt"
	"log"
	"net"
        "os"
)

// Logger is value capable of sending logs to the log server.
type Logger struct {
	conn *net.TCPConn
	node string
}

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
	log.SetPrefix(logger.node + " ")
	return
}

func (logger *Logger) Print(args ...interface{}) {
	fmt.Println(args...)
	log.Print(args...)
}

func (logger *Logger) Printf(format string, args ...interface{}) {
	log.Printf(format, args...)
	fmt.Printf(format+"\n", args...)
}

func (logger *Logger) Println(args ...interface{}) {
	fmt.Println(args...)
	log.Println(args...)
}

func (logger *Logger) Shutdown() {
	logger.conn.Close()
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
