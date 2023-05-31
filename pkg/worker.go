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
)

type msg struct {
	workerId int
	msgType int
	data    interface{}
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
			opts.result <- msg{workerId, opts.id, msgType: ERROR, data: r}
		}
	}()

	switch f := opts.fun.(type) {
	case func:

	}
}

func isFunc(i interface()) bool {
	return reflect.TypeOf(i).Kin() == reflect.Func
}