package easyworker

import (
	"fmt"
	"time"
)

func FullTest() {

	fn3 := func(a ...int) int {
		sum := 0
		for _, i := range a {
			sum += i
		}
		return sum
	}

	fn := func(a int, b int) int {
		if a%3 == 0 {
			panic("panic from user func")
		}
		return a + b
	}

	fn2 := func(a int, suffix string) string {
		if a%3 == 0 {
			panic("panic from user func")
		}
		return fmt.Sprintf("%d_%s", a, suffix)
	}

	eWorker, err := NewEasyWorker(fn3, 3, 1)
	if err != nil {
		fmt.Println("cannot create EasyWorker, ", err)
		return
	}

	eWorker.AddParams(1, 2, 3)
	eWorker.AddParams(3, 4, 5, 6, 7)

	r, e := eWorker.Run()
	if e != nil {
		fmt.Println("run task failed, ", e)
	} else {
		fmt.Println("task result:", r)
	}

	eWorker, err = NewEasyWorker(fn, 1, 0)
	if err != nil {
		fmt.Println("cannot create EasyWorker, ", err)
		return
	}

	for i := 1; i <= 5; i++ {
		eWorker.AddParams(i, i)
	}
	r, e = eWorker.Run()
	if e != nil {
		fmt.Println("run task failed, ", e)
	} else {
		fmt.Println("task result:", r)
	}

	inCh := make(chan []interface{}, 1)
	outCh := make(chan interface{})

	time.Sleep(time.Millisecond)

	// test with stream.
	eWorker, err = NewEasyWorkerStream(inCh, outCh, fn2, 2, 1)
	if err != nil {
		fmt.Println("create EasyWorker failed, ", err)
	}

	e = eWorker.RunStream()
	if e != nil {
		fmt.Println("run stream task failed, ", e)
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
