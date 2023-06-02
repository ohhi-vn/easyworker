package easyworker

import (
	"fmt"
	"testing"
	"time"
)

func Add(a int, b int) int {
	return a + b
}

func AddWithPanic(a int, b int) int {
	if a%3 == 0 {
		panic("panic from user func")
	}
	return a + b
}

func Sum(a ...int) int {
	sum := 0
	for _, i := range a {
		sum += i
	}
	return sum
}

func StrId(a int, suffix string) string {
	if a%3 == 0 {
		panic("panic from user func")
	}
	return fmt.Sprintf("%d_%s", a, suffix)
}

func TestIsNotFunc(t *testing.T) {
	_, err := NewTask("fn", 1, 0)
	if err == nil {
		t.Error("missed check function, ", err)
	}
}

func TestIncorrectNumWorker(t *testing.T) {
	_, err := NewTask(Add, 0, 0)
	if err == nil {
		t.Error("incorrect number of worker is passed, ", err)
	}
}

func TestIncorrectNumRetry(t *testing.T) {
	_, err := NewTask(Add, 1, -1)
	if err == nil {
		t.Error("incorrect number of retry is passed, ", err)
	}
}

func TestIsNotFuncStream(t *testing.T) {
	inCh := make(chan []interface{}, 1)
	outCh := make(chan interface{})

	// test with stream.
	_, err := NewStream(inCh, outCh, []int{1, 3}, 2, 1)
	if err == nil {
		t.Error("missed check function, ", err)
	}
}

func TestIncorrectNumWorkerStream(t *testing.T) {
	inCh := make(chan []interface{}, 1)
	outCh := make(chan interface{})

	// test with stream.
	_, err := NewStream(inCh, outCh, Add, -2, 0)
	if err == nil {
		t.Error("incorrect number of worker is passed, ", err)
	}
}

func TestIncorrectNumRetryStream(t *testing.T) {
	inCh := make(chan []interface{}, 1)
	outCh := make(chan interface{})

	// test with stream.
	_, err := NewStream(inCh, outCh, Add, 2, -1)
	if err == nil {
		t.Error("incorrect number of retry is passed, ", err)
	}
}

func TestNoTask(t *testing.T) {
	eWorker, _ := NewTask(Add, 0, 0)

	_, e := eWorker.Run()
	if e == nil {
		t.Error("easyworker run without task", e)
	}
}

func TestTaskList1(t *testing.T) {
	eWorker, err := NewTask(AddWithPanic, 1, 0)
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
		fmt.Println("task result:", r)
	}
}

func TestTaskList2(t *testing.T) {
	eWorker, err := NewTask(Sum, 3, 1)
	if err != nil {
		t.Error("cannot create EasyWorker, ", err)
		return
	}

	eWorker.AddTask(1, 2, 3)
	eWorker.AddTask(3, 4, 5, 6, 7)

	r, e := eWorker.Run()
	if e != nil {
		t.Error("run task failed, ", e)
	} else {
		fmt.Println("task result:", r)
	}

}

func TestStreamStopWithOutRun(t *testing.T) {
	inCh := make(chan []interface{}, 1)
	outCh := make(chan interface{})

	// test with stream.
	eWorker, err := NewStream(inCh, outCh, StrId, 2, 1)
	if err != nil {
		t.Error("create EasyWorker failed, ", err)
	}

	err = eWorker.StopStream()
	if err == nil {
		t.Error("StopStream malfunction")
	}
}

func TestStream(t *testing.T) {
	inCh := make(chan []interface{}, 1)
	outCh := make(chan interface{})

	time.Sleep(time.Millisecond)

	// test with stream.
	eWorker, err := NewStream(inCh, outCh, StrId, 2, 1)
	if err != nil {
		t.Error("create EasyWorker failed, ", err)
	}

	e := eWorker.Run()
	if e != nil {
		t.Error("run stream task failed, ", e)
	} else {
		fmt.Println("stream is running")
	}

	go func() {
		for {
			r := <-outCh
			fmt.Println("stream result: ", r)
		}
	}()

	go func() {
		for i := 0; i < 15; i++ {
			input := []interface{}{i, "3"}
			inCh <- input
			fmt.Println("stream sent: ", input)

		}
	}()

	time.Sleep(2 * time.Second)

	fmt.Println("send stop signal to stream")
	eWorker.StopStream()

	time.Sleep(time.Second)
}
