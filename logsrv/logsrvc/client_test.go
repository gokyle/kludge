package logsrvc

import (
	"flag"
	"fmt"
	"testing"
)

var (
	logSrv string
	logger *Logger
)

func init() {
	flAddr := flag.String("address", "127.0.0.1:5988", "address of log server")
	flag.Parse()
	logSrv = *flAddr
}

func TestConnect(t *testing.T) {
	var err error

	logger, err = Connect("test-client", logSrv)
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
