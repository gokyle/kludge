package common

const (
	OpGet = iota
	OpSet
	OpDel
	OpList
)

type Operation struct {
	OpCode byte
	Key    []byte
	Val    []byte
}

type Request struct {
	Op   *Operation
	Resp chan Response
}

type Response struct{}
