package easyworker

import (
	"testing"
	"time"
)

func TestStreamStopWithOutRun(t *testing.T) {
	inCh := make(chan []any, 1)
	outCh := make(chan any)

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
	inCh := make(chan []any, 1)
	outCh := make(chan any)

	time.Sleep(time.Millisecond)

	// test with stream.
	eWorker, err := NewStream(defaultConfig(StrId), inCh, outCh)
	if err != nil {
		t.Error("create EasyWorker failed, ", err)
	}

	e := eWorker.Run()
	if e != nil {
		t.Error("run stream task failed, ", e)
	}

	go func() {
		for {
			// get result from stream
			<-outCh
		}
	}()

	go func() {
		for i := 0; i < 15; i++ {
			input := []any{i, "3"}

			// send task to stream
			inCh <- input
		}
	}()

	time.Sleep(2 * time.Second)

	eWorker.Stop()

	time.Sleep(time.Second)
}
