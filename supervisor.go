package easyworker

import (
	"errors"
	"fmt"
)

var (
	// use to store last id of supervisor. id is auto_increment.
	superLastId int
)

/*
Config is shared between EasyTask & EasyStream struct.
It's store options that input by user.
*/
type Config struct {
	// id of supervisor.
	id int

	// fun stores function add by user.
	fun interface{}

	// number of workers (goroutines)
	worker int

	// retry times, if function was failed, worker will try again.
	retry int

	// sleep time before rerun
	retrySleep int
}

/*
Store options and runtime data for task processing.
Also, struct provides interface for control and processing task.
*/
type EasyTask struct {
	// config input by user.
	config Config

	// task for worker. It's slice of slice of params.
	inputs [][]interface{}

	// store runtime workers.
	workerList map[int]*worker
}

/*
Store options and runtime data for stream processing.
Also, struct provides interface for control and processing task.
*/
type EasyStream struct {
	// config input by user.
	config Config

	// inputs channel.
	inputCh chan []interface{}

	// output channel.
	outputCh chan interface{}

	// cmd channel for supervisor.
	cmdCh chan int

	// store runtime workers.
	workerList map[int]*worker
}

/*
Make a configuration holder for EasyTask or EasyStream.
fun: This is func you need to run task.
numWorkers: Number of goroutine you want to run task.
retryTimes: Number of retry if func is failed.

Example:

	fn = func(n int, prefix string) string {
		return fmt.Sprintf("%s_%d", prefix, n)
	}

	config,_ := NewConfig(fn, 3, 0, 0)
*/
func NewConfig(fun interface{}, numWorkers int, retryTimes int, retrySleep int) (ret Config, err error) {
	if err = verifyFunc(fun); err != nil {
		err = fmt.Errorf("not a function, you need give a real function")
		return
	}

	if numWorkers < 1 {
		err = fmt.Errorf("number of workers is incorrect, %d", numWorkers)
		return
	}

	if retryTimes < 0 {
		err = fmt.Errorf("retryTimes is incorrect, %d", numWorkers)
		return
	}

	if retryTimes > 0 && retrySleep < 0 {
		err = fmt.Errorf("retrySleep is incorrect, %d", numWorkers)
		return
	}

	ret = Config{
		id:         superLastId,
		fun:        fun,
		worker:     numWorkers,
		retry:      retryTimes,
		retrySleep: retrySleep,
	}

	return
}

/*
Make new EasyTask.
Config is made before make new EasyTask.

Example:

	task,_ := NewTask(config)
*/
func NewTask(config Config) (ret EasyTask, err error) {
	// auto incremental number, get supervisor's id/
	superLastId++

	ret = EasyTask{
		config:     config,
		inputs:     make([][]interface{}, 0),
		workerList: make(map[int]*worker, config.worker),
	}

	return
}

/*
Make new EasyStream.
Config is made before make new EasyTask.

taskCh: channel EasyStream will wait & get task.
resultCh: channel EastyStream will send out result of task.

Example:

	task,_ := NewStream(config)
*/
func NewStream(config Config, taskCh chan []interface{}, resultCh chan interface{}) (ret EasyStream, err error) {
	// auto incremental number, get supervisor's id/
	superLastId++

	ret = EasyStream{
		config:     config,
		inputCh:    taskCh,
		outputCh:   resultCh,
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

/*
Run func to process continuously.

Example:

	easyStream.Run()
*/
func (p *EasyStream) Run() (retErr error) {
	// use for send function's params to worker.
	inputCh := make(chan msg, p.config.worker)

	// use for get result from worker.
	resultCh := make(chan msg, p.config.worker)

	p.cmdCh = make(chan int)

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

		fmt.Println("stream start worker", i)
		go opt.run()
	}

	// Send data to worker
	go func() {
		for {
			params := <-p.inputCh
			fmt.Println("stream received new params: ", params)
			inputCh <- msg{id: iSTREAM, msgType: iTASK, data: params}
		}
	}()

	// receive result from worker
	go func() {
		for {
			result := <-resultCh
			switch result.msgType {
			case iSUCCESS: // task done
				fmt.Println("stream task", result.id, " is done, result:", result.data)
				p.outputCh <- result.data
			case iERROR: // task failed
				fmt.Println("stream task", result.id, " is failed, error:", result.data)
				p.outputCh <- result.data
			case iFATAL_ERROR: // worker panic
				fmt.Println(result.id, "worker (stream) is fatal error")
			case iQUIT: // worker quited
				fmt.Println(result.id, " exited (stream)")
			}
		}
	}()

	// send signal to worker to stop.
	go func() {
		for {
			cmd := <-p.cmdCh
			switch cmd {
			case iQUIT:
				for _, w := range p.workerList {
					w.cmd <- msg{msgType: iQUIT}
				}
			}
		}
	}()

	return
}

/*
Stop all workers.
*/
func (p *EasyStream) Stop() error {
	if p.cmdCh != nil {
		p.cmdCh <- iQUIT
		return nil
	} else {
		return errors.New("EasyWorker isn't sart or wrong task's type")
	}
}
