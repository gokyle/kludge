package main

import "encoding/json"
import "github.com/jmhodges/levigo"
import "github.com/gokyle/kludge/common"

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
			logger.Printf("worker %d returns", id)
			return
		}
		req.Op.WID = id

		logger.Printf("worker %d handling %s request", id,
			req.OpName())
		switch req.Op.OpCode {
		case common.OpGet:
			req.Resp <- store_get(req.Op)
		case common.OpSet:
			req.Resp <- store_set(req.Op)
		case common.OpDel:
			req.Resp <- store_del(req.Op)
		case common.OpLst:
			req.Resp <- store_lst(req.Op)
		default:
			logger.Printf("worker %d received invalid operation %d",
				id, req.Op.OpCode)
			resp := new(common.Response)
			resp.Body = []byte("invalid request")
			req.Resp <- resp
		}
	}
}

func store_get(op *common.Operation) (resp *common.Response) {
	ropts := levigo.NewReadOptions()
	ropts.SetVerifyChecksums(true)
	defer ropts.Close()
	resp = new(common.Response)

	data, err := ldb.Get(ropts, op.Key)
	if err != nil {
		logger.Printf("error handling get from worker %d: %s",
			op.WID, err.Error())
		resp.ErrMsg = err.Error()
	} else {
		logger.Printf("worker %d successfully completes GET", op.WID)
		resp.KeyOK = data != nil
		resp.Body = data
	}
	return
}

func store_set(op *common.Operation) (resp *common.Response) {
	resp = new(common.Response)
	ropts := levigo.NewReadOptions()
	ropts.SetVerifyChecksums(true)
	defer ropts.Close()

	data, err := ldb.Get(ropts, op.Key)
	if err != nil {
		logger.Printf("worker %d failed to read key: %s", op.WID,
			err.Error())
		resp.ErrMsg = err.Error()
		return
	}

	wopts := levigo.NewWriteOptions()
	wopts.SetSync(true)
	defer wopts.Close()

	err = ldb.Put(wopts, op.Key, op.Val)
	if err != nil {
		logger.Printf("worker %d failed to set key: %s", op.WID,
			err.Error())
		resp.ErrMsg = err.Error()
		return
	} else {
		logger.Printf("worker %d successfully wrote key", op.WID)
		resp.KeyOK = data != nil
		resp.Body = data
	}
	return
}

func store_del(op *common.Operation) (resp *common.Response) {
	ropts := levigo.NewReadOptions()
	ropts.SetVerifyChecksums(true)
	defer ropts.Close()

	resp = new(common.Response)
	data, err := ldb.Get(ropts, op.Key)
	if err != nil {
		logger.Printf("worker %d failed to read key: %s", op.WID,
			err.Error())
		resp.ErrMsg = err.Error()
		return
	}

	wopts := levigo.NewWriteOptions()
	wopts.SetSync(true)
	defer wopts.Close()

	err = ldb.Delete(wopts, op.Key)
	if err != nil {
		logger.Printf("worker %d failed to delete key: %s", op.WID,
			err.Error())
		resp.ErrMsg = err.Error()
		return
	} else {
		resp.KeyOK = data != nil
		resp.Body = data
	}
	return
}

func store_lst(op *common.Operation) (resp *common.Response) {
	resp = new(common.Response)
	keys := make([]string, 0)
	ropts := levigo.NewReadOptions()
	ropts.SetFillCache(true)
	defer ropts.Close()

	it := ldb.NewIterator(ropts)
	for it.SeekToFirst(); it.Valid(); it.Next() {
		keys = append(keys, string(it.Key()))
	}

	if err := it.GetError(); err != nil {
		logger.Printf("worker %d failed to iterate over keys: %s",
			op.WID, err.Error())
		resp.ErrMsg = err.Error()
	} else {
		resp.Body, err = json.Marshal(keys)
		if err != nil {
			logger.Printf("worker %d failed to create JSON response: %s",
				op.WID, err.Error())
			resp.Body = []byte{}
		}
	}
	return
}
