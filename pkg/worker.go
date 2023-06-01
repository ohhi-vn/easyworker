package main

import (
	"fmt"
	"reflect"
)

const (
	STREAM = -1
	ERROR  = iota
	FATAL_ERROR
	SUCCESS
	RETRY
	CANCEL
	RUN
	QUIT
	TASK
)

// struct carry task/command for worker.
type msg struct {
	id      int
	msgType int
	data    interface{}
}

// worker's information.
type worker struct {
	// worker's id
	id int

	// retry time, define by user.
	retryTime int

	// function, define by user.
	fun interface{}

	// command channel, supervisor uses to send command to worker.
	cmd chan msg

	// input channel, worker receives task (params) then run with fun.
	inputCh chan msg

	// output channel, worker send back result to supervisor.
	resultCh chan msg
}

/*
start worker with options in struct.
after start worker will wait task from supervisor.
after task done, worker will send result back to supervisor with id of task.
*/
func (opts *worker) run() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(opts.id, ", worker was panic, ", r)
			opts.resultCh <- msg{id: opts.id, msgType: FATAL_ERROR, data: r}
		}
	}()

	var (
		task msg
		ret  reflect.Value
		err  error
	)

	for {
		select {
		case task = <-opts.inputCh:
		case cmd := <-opts.cmd:
			// receive a quit signal.
			if cmd.msgType == QUIT {
				fmt.Println(opts.id, "is exited")
				return
			}
		}

		fmt.Println(opts.id, ", received new task, ", task, "data:", task.data)

		switch task.msgType {
		case TASK:
			args := task.data.([]interface{})
			i := 0
			for ; i <= opts.retryTime; i++ {
				if i > 0 {
					fmt.Println(opts.id, ", retry(", i, ") function with last args")
				}
				ret, err = InvokeFun(opts.fun, args...)
				if err == nil {
					break
				}
			}

			if err != nil {
				fmt.Println(opts.id, ", call function failed, error: ", err)
				opts.resultCh <- msg{id: task.id, msgType: ERROR, data: err}
			} else {
				fmt.Println(opts.id, ", function return ", ret)
				opts.resultCh <- msg{id: task.id, msgType: SUCCESS, data: ret}
			}
		}
	}
}

/*
call user's function througth reflect.
*/
func InvokeFun(fun interface{}, args ...interface{}) (ret reflect.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("call function failed, ", r)
			err = fmt.Errorf("call function failed, %s", r)
		}
	}()

	fmt.Println("list args: ", args)

	fn := reflect.ValueOf(fun)
	fnType := fn.Type()
	numIn := fnType.NumIn()
	if numIn > len(args) {
		return reflect.ValueOf(nil), fmt.Errorf("function must have minimum %d params. Have %d", numIn, len(args))
	}
	if numIn != len(args) && !fnType.IsVariadic() {
		return reflect.ValueOf(nil), fmt.Errorf("func must have %d params. Have %d", numIn, len(args))
	}
	in := make([]reflect.Value, len(args))
	for i := 0; i < len(args); i++ {
		var inType reflect.Type
		if fnType.IsVariadic() && i >= numIn-1 {
			inType = fnType.In(numIn - 1).Elem()
		} else {
			inType = fnType.In(i)
		}
		argValue := reflect.ValueOf(args[i])
		if !argValue.IsValid() {
			return reflect.ValueOf(nil), fmt.Errorf("func Param[%d] must be %s. Have %s", i, inType, argValue.String())
		}
		argType := argValue.Type()
		if argType.ConvertibleTo(inType) {
			in[i] = argValue.Convert(inType)
		} else {
			return reflect.ValueOf(nil), fmt.Errorf("method Param[%d] must be %s. Have %s", i, inType, argType)
		}
	}

	ret = fn.Call(in)[0]

	return ret, nil
}

/*
verify if interface is a function.
if interface is not a function, it will return an error.
*/
func verifyFunc(fun interface{}) error {
	if v := reflect.ValueOf(fun); v.Kind() != reflect.Func {
		return fmt.Errorf("you need give a real function")
	}
	return nil
}
