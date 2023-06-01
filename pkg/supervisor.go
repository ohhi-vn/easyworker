package main

import (
	"errors"
	"fmt"
)

const (
	TASK_STREAM = iota
	TASK_LIST
)

var (
	// use to store last id of supervisor. id is auto_increment.
	superLastId int
)

/*
EasyWorker is supervisor struct.
It's store options that input by user.
*/
type EasyWorker struct {
	// id of supervisor.
	id int

	// type of task: stream or existed tasks
	typeTask int

	// fun stores function add by user.
	fun interface{}

	// task for worker. It's slice of slice of params.
	inputs [][]interface{}

	// number of workers (goroutines)
	numWorker int

	// retry times, if function was failed, worker will try again.
	numRetry int

	// store runtime workers.
	workerList map[int]*worker

	// inputs channel
	inputCh chan []interface{}

	// output channel
	outputCh chan interface{}

	// cmd channel for supervisor
	cmdCh chan int
}

/*
Make new EasyWorker.
fun: This is func you need to run task.
numWorker: Number of goroutine you want to run task.
numRetry: Number of retry if func is failed.

Example:

	fn = func(n int, prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, n)
	}

	workers := NewEasyWoker(fn, 3, 0)
*/
func NewEasyWorker(fun interface{}, numWorker int, numRetry int) (ret EasyWorker, err error) {
	if err = verifyFunc(fun); err != nil {
		err = fmt.Errorf("not a function, you need give a real function")
		return
	}

	if numWorker < 1 {
		err = fmt.Errorf("number of workers is incorrect, %d", numWorker)
		return
	}

	if numWorker < 0 {
		err = fmt.Errorf("number of retry times is incorrect, %d", numWorker)
		return
	}

	// auto incremental number, get supervisor's id/
	superLastId++

	ret = EasyWorker{
		id:         superLastId,
		typeTask:   TASK_LIST,
		fun:        fun,
		inputs:     make([][]interface{}, 0),
		numWorker:  numWorker,
		numRetry:   numRetry,
		workerList: make(map[int]*worker, numWorker),
	}

	return
}

func NewEasyWorkerStream(taskCh chan []interface{}, resultCh chan interface{}, fun interface{}, numWorker int, numRetry int) (ret EasyWorker, err error) {
	if err = verifyFunc(fun); err != nil {
		err = fmt.Errorf("not a function, you need give a real function")
		return
	}

	if numWorker < 1 {
		err = fmt.Errorf("number of workers is incorrect, %d", numWorker)
		return
	}

	if numWorker < 0 {
		err = fmt.Errorf("number of retry times is incorrect, %d", numWorker)
		return
	}

	// auto incremental number, get supervisor's id/
	superLastId++

	ret = EasyWorker{
		id:         superLastId,
		typeTask:   TASK_STREAM,
		fun:        fun,
		inputs:     make([][]interface{}, 0),
		numWorker:  numWorker,
		numRetry:   numRetry,
		workerList: make(map[int]*worker, numWorker),
		inputCh:    taskCh,
		outputCh:   resultCh,
	}

	return
}

/*
Uses for adding tasks for workers.

Example:

	workers.AddParams(1, "user")
	workers.AddParams(2, "user")
	workers.AddParams(1000, "admin")
*/
func (p *EasyWorker) AddParams(i ...interface{}) {
	params := make([]interface{}, 0)
	params = append(params, i...)

	p.inputs = append(p.inputs, params)
}

/*
Run func with task.

Example:

	workers.Run()
*/
func (p *EasyWorker) Run() (ret []interface{}, retErr error) {
	ret = make([]interface{}, 0)

	fmt.Println("len:", len(p.inputs))
	fmt.Println("inputs: ", p.inputs)

	if len(p.inputs) < 1 {
		retErr = errors.New("need params to run")
		return
	}

	// use for send function's params to worker.
	inputCh := make(chan msg, p.numWorker)

	// use for get result from worker.
	resultCh := make(chan msg, p.numWorker)

	// Start workers
	for i := 0; i < p.numWorker; i++ {
		opt := &worker{
			id:        i,
			fun:       p.fun,
			cmd:       make(chan msg),
			resultCh:  resultCh,
			inputCh:   inputCh,
			retryTime: p.numRetry,
		}
		p.workerList[i] = opt

		fmt.Println("start worker", i)
		go opt.run()
	}

	// Send data to worker
	go func() {
		for index, params := range p.inputs {
			fmt.Println("send params: ", params)
			inputCh <- msg{id: index, msgType: TASK, data: params}
		}
	}()

	resultMap := map[int]interface{}{}

	// receive result from worker
	for {
		result := <-resultCh
		switch result.msgType {
		case SUCCESS: // task done
			fmt.Println("task", result.id, " is done, result:", result.data)
			resultMap[result.id] = result.data
		case ERROR: // task failed
			fmt.Println("task", result.id, " is failed, error:", result.data)
			resultMap[result.id] = result.data
		case FATAL_ERROR: // worker panic
			fmt.Println(result.id, "worker is fatal error")
		case QUIT: // worker quited
			fmt.Println(result.id, " exited")
		}

		if len(resultMap) == len(p.inputs) {
			fmt.Println("collect all result, ", resultMap)
			break
		}
	}

	// send signal to worker to stop.
	go func() {
		for _, w := range p.workerList {
			w.cmd <- msg{msgType: QUIT}
		}
	}()

	ret = make([]interface{}, len(resultMap))

	for k, v := range resultMap {

		ret[k] = v
	}

	return
}

/*
Run func with task.

Example:

	workers.Run()
*/
func (p *EasyWorker) RunStream() (retErr error) {
	fmt.Println("len:", len(p.inputs))
	fmt.Println("inputs: ", p.inputs)

	// use for send function's params to worker.
	inputCh := make(chan msg, p.numWorker)

	// use for get result from worker.
	resultCh := make(chan msg, p.numWorker)

	p.cmdCh = make(chan int)

	// Start workers
	for i := 0; i < p.numWorker; i++ {
		opt := &worker{
			id:        i,
			fun:       p.fun,
			cmd:       make(chan msg),
			resultCh:  resultCh,
			inputCh:   inputCh,
			retryTime: p.numRetry,
		}
		p.workerList[i] = opt

		fmt.Println("stream start worker", i)
		go opt.run()
	}

	// Send data to worker
	go func() {
		for {
			params := <-p.inputCh
			fmt.Println("stream received new params: ", params)
			inputCh <- msg{id: STREAM, msgType: TASK, data: params}
		}
	}()

	// receive result from worker
	go func() {
		for {
			result := <-resultCh
			switch result.msgType {
			case SUCCESS: // task done
				fmt.Println("stream task", result.id, " is done, result:", result.data)
				p.outputCh <- result.data
			case ERROR: // task failed
				fmt.Println("stream task", result.id, " is failed, error:", result.data)
				p.outputCh <- result.data
			case FATAL_ERROR: // worker panic
				fmt.Println(result.id, "worker (stream) is fatal error")
			case QUIT: // worker quited
				fmt.Println(result.id, " exited (stream)")
			}
		}
	}()

	// send signal to worker to stop.
	go func() {
		for {
			cmd := <-p.cmdCh
			switch cmd {
			case QUIT:
				for _, w := range p.workerList {
					w.cmd <- msg{msgType: QUIT}
				}
			}
		}
	}()

	return
}

func (p *EasyWorker) StopStream() error {
	if p.cmdCh != nil {
		p.cmdCh <- QUIT
		return nil
	} else {
		return errors.New("EasyWorker isn't sart or wrong task's type")
	}
}
