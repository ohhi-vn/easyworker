package main

import (
	"errors"
	"fmt"
)

var (
	cmdChan map[int]chan msg
)

func StartSuper(data []interface{}, fun interface{}, numWorker int) (in []interface{}, ret []interface{}, retErr error) {
	fmt.Println("%v", data)
	inputs := data //.([][]interface{})
	if len(inputs) < 1 || numWorker < 1 {
		retErr = errors.New("incorrect params")
		return
	}

	workerList := make(map[int]*worker, numWorker)

	// Start workers
	i := 1
	//	for i := 0; i < numWorker; i++ {
	opt := &worker{
		id:     i,
		fun:    fun,
		cmd:    make(chan msg),
		result: make(chan msg, numWorker),
	}
	workerList[i] = opt

	fmt.Println("start worker", i)
	go runWorker(opt)
	//	}

	// Send data to worker
	for d := range inputs {
		opt.cmd <- msg{msgType: RUN, data: d}
	}

	// Wait result

	return
}

func StartSuperStream(dataCh chan []interface{}, resultCh chan []interface{}, fun interface{}, numWorker int) (retErr error) {
	if e := verifyFunc(fun); e != nil {
		return e
	}
	workerList := make(map[int]*worker, numWorker)

	// Start workers
	i := 1
	//	for i := 0; i < numWorker; i++ {
	opt := &worker{
		id:     i,
		fun:    fun,
		cmd:    make(chan msg),
		result: make(chan msg, numWorker),
	}
	workerList[i] = opt

	fmt.Println("start worker", i)
	go runWorker(opt)
	//	}

	// Send data to worker
	go func() {
		for {
			d := <-dataCh
			fmt.Println("supervisor get a new task, ", d)
			opt.cmd <- msg{msgType: RUN, data: d}
		}
	}()
	// Wait result

	return nil
}
