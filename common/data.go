package common

const (
	OpGet = iota
	OpSet
	OpDel
	OpLst
)

var opNames map[int]string

func init() {
	opNames = make(map[int]string, 0)
	opNames[OpGet] = "GET"
	opNames[OpSet] = "SET"
	opNames[OpDel] = "DEL"
	opNames[OpLst] = "LST"
}

type Operation struct {
	OpCode byte
	Key    []byte
	Val    []byte
        WID    int      // ID of the handling worker
}

func (op *Operation) Name() string {
	return opNames[op.OpCode]
}

type Request struct {
	Op   *Operation
	Resp chan Response
}

func (req *Request) OpName() string {
	return req.Op.Name()
}

type Response struct {
	Body []byte
}
