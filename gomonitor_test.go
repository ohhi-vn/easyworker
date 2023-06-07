package easyworker

import (
	"fmt"
	"testing"
	"time"
)

func TestGoRun(t *testing.T) {
	g, err := NewGo(loopRun2, 5)

	if err != nil {
		t.Error("create go failed, ", err)
		return
	}

	refId, ch := g.Monitor()

	g.Run()

	select {
	case sig := <-ch:
		if refId != sig.RefId || SIGNAL_DONE != sig.Signal {
			t.Error("return signal incorrect")
		}
	case <-time.After(time.Second):
		t.Error("timed out")
	}
}

func TestGoMultiMonitor(t *testing.T) {
	g, err := NewGo(loopRun2, 5)

	if err != nil {
		t.Error("create go failed, ", err)
		return
	}

	okCh := make(chan bool)

	go func(out chan<- bool) {
		refId, ch := g.Monitor()

		sig := <-ch
		if refId != sig.RefId || SIGNAL_DONE != sig.Signal {
			out <- false
		} else {
			fmt.Println("ref:", refId, "ok")
			out <- true
		}
	}(okCh)

	go func(out chan<- bool) {
		refId, ch := g.Monitor()

		sig := <-ch
		if refId != sig.RefId || SIGNAL_DONE != sig.Signal {
			out <- false
		} else {
			fmt.Println("ref:", refId, "ok")
			out <- true
		}
	}(okCh)

	g.Run()

	for i := 0; i < 2; i++ {
		select {
		case ret := <-okCh:
			if !ret {
				t.Error("return signal incorrect")
				return
			}
		case <-time.After(time.Second):
			t.Error("timed out")
			return
		}
	}
}
