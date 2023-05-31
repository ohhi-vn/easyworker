package main

import (
	"fmt"
	"reflect"
)

const (
	ERROR = iota
	SUCCESS
	RETRY
	CANCEL
	RUN
)

type msg struct {
	workerId int
	msgType  int
	data     interface{}
}

type worker struct {
	id        int
	isDone    bool
	retryTime int
	fun       interface{}
	cmd       chan msg
	result    chan msg
}

func Hello() {
	fmt.Println("test")
}

func runWorker(opts *worker) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("worker", opts.id, " was panic, ", r)
			opts.result <- msg{workerId: opts.id, msgType: ERROR, data: r}
		}
	}()

	for {
		task := <-opts.cmd

		fmt.Println(opts.id, ", received new task, ", task)

		switch task.msgType {
		case RUN:
			args := task.data.([]interface{})
			r, e := InvokeFun(opts.fun, args...)
			fmt.Println("result: ", r, ", error: ", e)

		}
	}
}

func InvokeFun(fun interface{}, args ...interface{}) (reflect.Value, error) {
	fmt.Println("list args: ", args)
	if v := reflect.ValueOf(fun); v.Kind() != reflect.Func {
		return reflect.ValueOf(nil), fmt.Errorf("you need give a real function")
	}
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
	return fn.Call(in)[0], nil
}

func verifyFunc(fun interface{}) error {
	if v := reflect.ValueOf(fun); v.Kind() != reflect.Func {
		return fmt.Errorf("you need give a real function")
	}
	return nil
}
