package easyworker

import (
	"fmt"
	"testing"
	"time"
)

func TestStreamStopWithOutRun(t *testing.T) {
	inCh := make(chan []interface{}, 1)
	outCh := make(chan interface{})

	// test with stream.
	eWorker, err := NewStream(defaultConfig(StrId), inCh, outCh)
	if err != nil {
		t.Error("create EasyWorker failed, ", err)
	}

	err = eWorker.Stop()
	if err == nil {
		t.Error("StopStream malfunction")
	}
}

func TestStream(t *testing.T) {
	inCh := make(chan []interface{}, 1)
	outCh := make(chan interface{})

	time.Sleep(time.Millisecond)

	// test with stream.
	eWorker, err := NewStream(defaultConfig(StrId), inCh, outCh)
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
	eWorker.Stop()

	time.Sleep(time.Second)
}
