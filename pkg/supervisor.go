package main

import (
	"errors"
	"fmt"
)

var (
	cmdChan map[int]chan msg
)

type userFunc func(a int, b int) int

func StartSuper(data [][]interface{}, fun interface{}, numWorker int) (ret []interface{}, retErr error) {
	if len(data) < 1 || numWorker < 1 {
		retErr = errors.New("incorrect params")
		return
	}

	workerList := make(map[int]*worker, numWorker)

	// Start workers
	for i := 0; i < numWorker; i++ {
		opt := &worker{
			id:     i,
			fun:    fun,
			cmd:    make(chan msg),
			result: make(chan msg, numWorker),
		}
		workerList[i] = opt

		fmt.Println("start worker", i)
		go runWorker(opt)
	}

	// Send data to worker

	// Wait result
}
