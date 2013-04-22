package main

import "github.com/jmhodges/levigo"
import "github.com/gokyle/kludge/common"
import "log"

var reqQ chan *common.Request

func startPool() {
	reqQ = make(chan *common.Request, reqBuf)
	for i := 0; i < poolSize; i++ {
		go requestHandler(i)
	}
}

func requestHandler(id int) {
	for {
		req, ok := <-reqQ
		if !ok {
			log.Printf("worker %d returns", id)
			return
		}
		req.Op.WID = id

		log.Printf("worker %d handling %s request", id,
			req.OpName())
		switch req.Op.OpCode {
		case common.OpGet:
			req.Resp <- store_get(req.Op)
		case common.OpSet:
			req.Resp <- store_set(req.Op)
		default:
			log.Printf("worker %d received invalid operation %d",
				id, req.Op.OpCode)
		}
	}
}

func store_get(op *common.Operation) (resp *common.Response) {
	ropts := levigo.NewReadOptions()
	ropts.SetVerifyChecksums(true)

	data, err := ldb.Get(ropts, op.Key)
	if err != nil {
		log.Printf("error handling get from worker %d: %s",
			op.WID, err.Error())
	} else {
		log.Printf("worker %d successfully completes GET", op.WID)
		resp = new(common.Response)
		resp.Body = data
	}
	return
}

func store_set(op *common.Operation) (resp *common.Response) {
	ropts := levigo.NewReadOptions()
	ropts.SetVerifyChecksums(true)

	data, err := ldb.Get(ropts, op.Key)
	if err != nil {
		log.Printf("worker %d failed to read key: %s", op.WID,
			err.Error())
		return
	}

	wopts := levigo.NewWriteOptions()
	wopts.SetSync(true)
	err = ldb.Put(wopts, op.Key, op.Val)
	if err != nil {
		log.Printf("worker %d failed to set key: %s", op.WID,
			err.Error())
		return
	} else {
		log.Printf("worker %d successfully wrote key", op.WID)
		resp = new(common.Response)
		resp.Body = data
	}
	return
}
