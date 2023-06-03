package easyworker

import (
	"fmt"
	"reflect"
)

const (
	iSTREAM = -1

	iERROR = iota
	iFATAL_ERROR
	iSUCCESS
	iRETRY
	iCANCEL
	iRUN
	iQUIT
	iTASK
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
	retryTimes int

	// sleep time between re-run.
	retrySleep int

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
func (w *worker) run() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(w.id, ", worker was panic, ", r)
			w.resultCh <- msg{id: w.id, msgType: iFATAL_ERROR, data: r}
		}
	}()

	var (
		task msg
		ret  reflect.Value
		err  error
	)

	for {
		select {
		case task = <-w.inputCh:
		case cmd := <-w.cmd:
			// receive a quit signal.
			if cmd.msgType == iQUIT {
				fmt.Println(w.id, "is exited")
				return
			}
		}

		fmt.Println(w.id, ", received new task, ", task, "data:", task.data)

		switch task.msgType {
		case iTASK:
			args := task.data.([]interface{})

			for i := 0; i <= w.retryTimes; i++ {
				if i > 0 {
					fmt.Println(w.id, ", retry(", i, ") function with last args")
				}
				ret, err = invokeFun(w.fun, args...)
				if err == nil {
					break
				}
			}

			if err != nil {
				fmt.Println(w.id, ", call function failed, error: ", err)
				w.resultCh <- msg{id: task.id, msgType: iERROR, data: err}
			} else {
				fmt.Println(w.id, ", function return ", ret)
				w.resultCh <- msg{id: task.id, msgType: iSUCCESS, data: ret}
			}
		}
	}
}

/*
call user's function througth reflect.
*/
func invokeFun(fun interface{}, args ...interface{}) (ret reflect.Value, err error) {
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
	params := make([]reflect.Value, len(args))
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
			params[i] = argValue.Convert(inType)
		} else {
			return reflect.ValueOf(nil), fmt.Errorf("method Param[%d] must be %s. Have %s", i, inType, argType)
		}
	}

	result := fn.Call(params)

	fmt.Println("invoke result:", result)

	if len(result) > 0 {
		ret = result[0]
	}

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
