package easyworker

import (
	"log"
	"testing"
)

func TestNoTask(t *testing.T) {
	config := defaultConfig(add)

	eWorker, _ := NewTask(config)

	_, e := eWorker.Run()
	if e == nil {
		t.Error("easyworker run without task", e)
	}
}

func TestTaskList1(t *testing.T) {
	eWorker, err := NewTask(defaultConfig(addWithPanic))
	if err != nil {
		t.Error("cannot create EasyWorker, ", err)
		return
	}

	for i := 1; i <= 5; i++ {
		eWorker.AddTask(i, i)
	}
	r, e := eWorker.Run()
	if e != nil {
		t.Error("run task failed, ", e)
	} else {
		log.Println("task result:", r)
	}
}

func TestTaskList2(t *testing.T) {
	eWorker, err := NewTask(defaultConfig(sum))
	if err != nil {
		t.Error("cannot create EasyWorker, ", err)
		return
	}

	eWorker.AddTask(1, 2, 3)
	eWorker.AddTask(3, 4, 5, 6, 7)

	_, e := eWorker.Run()
	if e != nil {
		t.Error("run task failed, ", e)
	}
}
