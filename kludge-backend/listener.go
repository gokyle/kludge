package main

import (
	"encoding/gob"
	"log"
	"net"
)

type operation struct {
	OpCode byte
	Key    []byte
	Val    []byte
}

func listener() {
	addr, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		log.Fatal("failed to resolve TCP address: ", err.Error())
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %s\n", listenAddr,
			err.Error())
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("listener failure: ", err.Error())
			continue
		}
		log.Print("connection: ", conn)
	}
}

func receiver(conn *net.TCPConn) {
	dec := gob.NewDecoder(conn)

	var op = new(operation)
	dec.Decode(op)
	reqQ <- op
}
