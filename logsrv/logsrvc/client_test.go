package logsrvc

import (
	"flag"
	"fmt"
	"github.com/gokyle/uuid"
	"os"
	"testing"
)

var (
	logSrv string
	logger *Logger
	node   string
)

func init() {
	flAddr := flag.String("address", "127.0.0.1:5988", "address of log server")
	flag.Parse()
	logSrv = *flAddr

	var err error
	node, err = uuid.GenerateV4String()
	if err != nil {
		fmt.Println("couldn't generate UUID:", err.Error())
		os.Exit(1)
	}
}

func TestConnect(t *testing.T) {
	var err error

	logger, err = Connect(node, logSrv)
	if err != nil {
		fmt.Printf("[!] error setting up test client: %s\n", err.Error())
		t.FailNow()
	}
}

func TestPrint(t *testing.T) {
	logger.Print("hello, world")
}

func TestPrintf(t *testing.T) {
	logger.Printf("testing log server %s", logSrv)
}

func TestShutdown(t *testing.T) {
	logger.Shutdown()
}
