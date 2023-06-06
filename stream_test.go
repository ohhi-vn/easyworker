package easyworker

import (
	"testing"
	"time"
)

func TestStreamStopWithOutRun(t *testing.T) {
	inCh := make(chan []any, 1)
	outCh := make(chan any)

	// test with stream.
	eWorker, err := NewStream(defaultConfig(strId), inCh, outCh)
	if err != nil {
		t.Error("create EasyWorker failed, ", err)
	}

	err = eWorker.Stop()
	if err == nil {
		t.Error("StopStream malfunction")
	}
}

func TestStream(t *testing.T) {
	inCh := make(chan []any, 1)
	outCh := make(chan any)

	// test with stream.
	eWorker, err := NewStream(defaultConfig(strId), inCh, outCh)
	if err != nil {
		t.Error("create EasyStream failed, ", err)
	}

	e := eWorker.Run()
	if e != nil {
		t.Error("run stream task failed, ", e)
	}

	num := 5

	go func() {
		for i := 0; i < num; i++ {
			input := []any{i, "hello"}

			// send task to stream
			inCh <- input
		}
	}()

	counterCh := make(chan int)

	go func() {
		counter := 0
	l:
		for {
			select {
			// get result from stream
			case <-outCh:
				counter++
			case <-time.After(time.Millisecond * 200):
				break l
			}
		}
		counterCh <- counter
	}()

	out := <-counterCh

	if out != num {
		t.Error("wrong result")
	}

	eWorker.Stop()
}
