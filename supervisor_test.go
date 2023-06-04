package easyworker

import (
	"testing"
	"time"
)

func LoopRun(a int, testSupporter chan int) {
	testSupporter <- a
	for i := 0; i < a; i++ {
		time.Sleep(time.Millisecond)
	}
}

func LoopRunWithPanic(a int, testSupporter chan int) {
	testSupporter <- a
	for i := 0; i < a; i++ {
		time.Sleep(time.Millisecond)
		if i == 1 {
			panic("test loop with panic")
		}
	}
}

func TestSupAlwaysRestart1(t *testing.T) {
	ch := make(chan int)

	sup := NewSupervisor()

	child, _ := NewChild(ALWAYS_RESTART, LoopRun, 5, ch)

	sup.AddChild(&child)

	counter := 0
l:
	for {
		select {
		case <-ch:
			counter++
			if counter > 3 {
				break l
			}
		case <-time.After(time.Second):
			t.Error("timed out")
			break l
		}
	}
}

func TestSupAlwaysRestart2(t *testing.T) {
	ch := make(chan int)

	sup := NewSupervisor()

	child, _ := NewChild(ALWAYS_RESTART, LoopRunWithPanic, 5, ch)

	sup.AddChild(&child)

	counter := 0
l:
	for {
		select {
		case <-ch:
			counter++
			if counter > 3 {
				break l
			}
		case <-time.After(3 * time.Second):
			t.Error("timed out")
		}
	}
}

func TestSupNormalRestart1(t *testing.T) {
	ch := make(chan int)

	sup := NewSupervisor()

	child, _ := NewChild(ERROR_RESTART, LoopRun, 5, ch)

	sup.AddChild(&child)

	counter := 0
l:
	for {
		select {
		case <-ch:
			counter++
			if counter > 1 {
				t.Error("unexpected, child was restarted in ERROR_RESTART strategy, fun run sucessful")
			}
		case <-time.After(time.Second):
			break l
		}
	}
}

func TestSupNormalRestart2(t *testing.T) {
	ch := make(chan int)

	sup := NewSupervisor()

	child, _ := NewChild(ERROR_RESTART, LoopRunWithPanic, 5, ch)

	sup.AddChild(&child)

	counter := 0
l:
	for {
		select {
		case <-ch:
			counter++
			if counter > 3 {
				break l
			}
		case <-time.After(time.Second):
			t.Error("timed out")
		}
	}
}

func TestSupStop(t *testing.T) {
	ch := make(chan int)

	sup := NewSupervisor()

	sup.NewChild(ALWAYS_RESTART, LoopRun, 3, ch)
	sup.NewChild(ALWAYS_RESTART, LoopRun, 3, ch)

	counter := 0
l:
	for {
		<-ch
		counter++
		if counter > 5 {
			sup.Stop()
			break l
		}
	}

l2:
	for {
		select {
		case <-ch:
		case <-time.After(time.Second):
			break l2
		}
	}

	for _, child := range sup.children {
		if child.canRun() {
			t.Error("stop supervisor failed")
			break
		}
	}
}
