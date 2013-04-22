package main

import "log"

var reqQ chan *operation

func startPool() {
	reqQ = make(chan *operation, reqBuf)
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
		switch req.OpCode {

		}
	}
}
