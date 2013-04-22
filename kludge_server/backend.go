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

func sendRequest(req *common.Operation) (data []byte, err error) {
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

func getKey(key string) ([]byte, error) {
        op := &common.Operation{
                OpCode: common.OpGet,
                Key: []byte(key),
            }
        return sendRequest(op)
}

func setKey(key string, value []byte) ([]byte, error) {
        op := &common.Operation{
                OpCode: common.OpSet,
                Key: []byte(key),
                Val: value,
        }
        return sendRequest(op)
}

func delKey(key string) ([]byte, error) {
        op := &common.Operation{
                OpCode: common.OpDel,
                Key: []byte(key),
            }
        return sendRequest(op)
}

func listKeys() ([]byte, error) {
        return sendRequest(&common.Operation{
                OpCode: common.OpLst,
        })
}
