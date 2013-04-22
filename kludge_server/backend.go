package main

import (
	"encoding/gob"
	"github.com/gokyle/kludge/common"
	"log"
	"net"
)

var nodeAddr *net.TCPAddr

func init() {
	var err error
	nodeAddr, err = net.ResolveTCPAddr("tcp", "127.0.0.1:5987")
	if err != nil {
		log.Fatal("failed to resolve TCP address: ", err.Error())
	}
}

func newGet(key string) (op *common.Operation) {
	op = new(common.Operation)
	op = new(common.Operation)
	op.OpCode = common.OpGet
	op.Key = []byte(key)
	return
}

func getKey(key string) (data []byte, err error) {
	req := newGet(key)
	conn, err := net.DialTCP("tcp", nil, nodeAddr)
	if err != nil {
		log.Println("TCP connection failed: ", err.Error())
		return
	}
	enc := gob.NewEncoder(conn)
	dec := gob.NewDecoder(conn)
	resp := new(common.Response)
	if err = enc.Encode(req); err != nil {
		log.Print("failed to encode request: ", err.Error())
		return
	}
	if err = dec.Decode(resp); err != nil {
		log.Printf("failed to decode response: ", err.Error())
		return
	}
	data = resp.Body
	return
}
