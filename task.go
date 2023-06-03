package easyworker

import (
	"errors"
	"fmt"
)

/*
Store options and runtime data for task processing.
Also, struct provides interface for control and processing task.
*/
type EasyTask struct {
	id int

	// config input by user.
	config Config

	// task for worker. It's slice of slice of params.
	inputs [][]interface{}

	// store runtime workers.
	workerList map[int]*worker
}

/*
Make new EasyTask.
Config is made before make new EasyTask.

Example:

	task,_ := NewTask(config)
*/
func NewTask(config Config) (ret EasyTask, err error) {
	// auto incremental number.
	taskLastId++

	ret = EasyTask{
		id:         taskLastId,
		config:     config,
		inputs:     make([][]interface{}, 0),
		workerList: make(map[int]*worker, config.worker),
	}

	return
}

/*
Uses for adding tasks for EasyTask.

Example:

	workers.AddParams(1, "user")
	workers.AddParams(2, "user")
	workers.AddParams(1000, "admin")
*/
func (p *EasyTask) AddTask(i ...interface{}) {
	params := make([]interface{}, 0)
	params = append(params, i...)

	p.inputs = append(p.inputs, params)
}

/*
Run func with existed task or waiting a new task.

Example:

	easyTask.Run()
*/
func (p *EasyTask) Run() (ret []interface{}, retErr error) {
	ret = make([]interface{}, 0)

	fmt.Println("len:", len(p.inputs))
	fmt.Println("inputs: ", p.inputs)

	if len(p.inputs) < 1 {
		retErr = errors.New("need params to run")
		return
	}

	// use for send function's params to worker.
	inputCh := make(chan msg, p.config.worker)

	// use for get result from worker.
	resultCh := make(chan msg, p.config.worker)

	// Start workers
	for i := 0; i < p.config.worker; i++ {
		opt := &worker{
			id:         i,
			fun:        p.config.fun,
			cmd:        make(chan msg),
			resultCh:   resultCh,
			inputCh:    inputCh,
			retryTimes: p.config.retry,
		}
		p.workerList[i] = opt

		fmt.Println("start worker", i)
		go opt.run()
	}

	// Send data to worker
	go func() {
		for index, params := range p.inputs {
			fmt.Println("send params: ", params)
			inputCh <- msg{id: index, msgType: iTASK, data: params}
		}
	}()

	resultMap := map[int]interface{}{}

	// receive result from worker
	for {
		result := <-resultCh
		switch result.msgType {
		case iSUCCESS: // task done
			fmt.Println("task", result.id, " is done, result:", result.data)
			resultMap[result.id] = result.data
		case iERROR: // task failed
			fmt.Println("task", result.id, " is failed, error:", result.data)
			resultMap[result.id] = result.data
		case iFATAL_ERROR: // worker panic
			fmt.Println(result.id, "worker is fatal error")
		case iQUIT: // worker quited
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
			w.cmd <- msg{msgType: iQUIT}
		}
	}()

	ret = make([]interface{}, len(resultMap))

	for k, v := range resultMap {

		ret[k] = v
	}

	return
}
