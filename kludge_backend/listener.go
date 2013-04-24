package main

import (
	"encoding/gob"
	"github.com/gokyle/kludge/common"
	"net"
	"time"
)

func listener() {
	addr, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		logger.Fatal("failed to resolve TCP address: ", err.Error())
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		logger.Fatalf("failed to listen on %s: %s\n", listenAddr,
			err.Error())
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Println("listener failure: ", err.Error())
			continue
		}
		go receiver(conn)
	}
}

func receiver(conn net.Conn) {
	defer conn.Close()
	start := time.Now().UnixNano()
	req := new(common.Request)
	dec := gob.NewDecoder(conn)
	respc := make(chan *common.Response)

	var op = new(common.Operation)
	dec.Decode(op)
	req.Op = op
	req.Resp = respc
	reqQ <- req

	resp := <-respc
	defer close(respc)

	enc := gob.NewEncoder(conn)
	enc.Encode(resp)
	rtime := (time.Now().UnixNano() - start) / 1000.0
	logger.Printf("%s response time: %dus", op.Name(), rtime)
}
