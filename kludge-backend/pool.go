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

		log.Printf("worker %d handling %s request", id,
			req.OpName())
		switch req.Op.OpCode {
		case common.OpGet:
                        resp := store_get(req.Op)
		}
	}
}

func store_get(op common.Request) (resp common.Response) {
	ropts := levigo.NewReadOptions()
        ropts.SetVerifyChecksums(true)

        data, err := ldb.Get(ropts, op.Key)
        if err != nil {
                log.Printf("error handling get from worker %d: %s",
                        id, err.Error())
        } else {
                log.Printf("worker %d successfully completes GET", id)
                resp = new(common.Response)
                resp.Body = data
        }
        return
}
