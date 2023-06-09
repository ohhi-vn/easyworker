package easyworker

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"
)

func simpleLoop(a int) {
	for i := 0; i < a; i++ {
		time.Sleep(time.Millisecond)
	}
}

func simpleLoopNoArg() {
	for i := 0; i < 50; i++ {
		time.Sleep(time.Millisecond)
	}
}

func loopRun(a int, testSupporter chan int) {
	testSupporter <- a
	for i := 0; i < a; i++ {
		time.Sleep(time.Millisecond)
	}
}

func simpleLoopWithPanic(a int) {
	for i := 0; i < a; i++ {
		time.Sleep(time.Millisecond)
		if i == 1 {
			panic("test loop with panic")
		}
	}
}

func simpleLoopWithContext(ctx context.Context, a int) {
	if supId := ctx.Value(CTX_SUP_ID); supId != nil {
		log.Println("Id of supervisor get from user function:", supId)
	} else {
		panic("context failed")
	}

	if childId := ctx.Value(CTX_CHILD_ID); childId != nil {
		log.Println("Id of child get from user function:", childId)
	} else {
		panic("context failed")
	}

	for i := 0; i < a; i++ {
		time.Sleep(time.Millisecond)
	}
}

func loopRunWithPanic(a int, testSupporter chan int) {
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

	child, _ := NewChild(ALWAYS_RESTART, loopRun, 5, ch)

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

	child, _ := NewChild(ALWAYS_RESTART, loopRunWithPanic, 5, ch)

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

func TestSupNormalRestart1(t *testing.T) {
	ch := make(chan int)

	sup := NewSupervisor()

	child, _ := NewChild(ERROR_RESTART, loopRun, 5, ch)

	sup.AddChild(&child)

	counter := 0
l:
	for {
		select {
		case <-ch:
			counter++
			if counter > 1 {
				t.Error("unexpected, child was restarted (task done) in ERROR_RESTART strategy")
			}
		case <-time.After(time.Second):
			break l
		}
	}
}

func TestSupNormalRestart2(t *testing.T) {
	ch := make(chan int)

	sup := NewSupervisor()

	child, _ := NewChild(ERROR_RESTART, loopRunWithPanic, 5, ch)

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

func TestSupNoRestart(t *testing.T) {
	sup := NewSupervisor()

	sup.NewChild(ALWAYS_RESTART, simpleLoop, 3)
	sup.NewChild(NO_RESTART, simpleLoop, 3)

	time.Sleep(time.Millisecond * 100)

	_, _, stopped, _ := sup.Stats()

	if stopped != 1 {
		t.Error("stop supervisor failed")
	}

	sup.Stop()
}

func TestSupFunNoArg(t *testing.T) {
	sup := NewSupervisor()

	sup.NewChild(ALWAYS_RESTART, simpleLoopNoArg)

	time.Sleep(time.Millisecond)

	_, running, _, _ := sup.Stats()

	if running != 1 {
		t.Error("start fun no arg failed")
	}

	sup.Stop()
}

func TestSupStop(t *testing.T) {
	ch := make(chan int)

	sup := NewSupervisor()

	sup.NewChild(ALWAYS_RESTART, loopRun, 3, ch)
	sup.NewChild(ALWAYS_RESTART, loopRun, 3, ch)

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

	total, _, stopped, _ := sup.Stats()

	if total != stopped {
		t.Error("stop supervisor failed")
	}

	sup.Stop()
}

func TestSupStopChild(t *testing.T) {
	sup := NewSupervisor()

	sup.NewChild(ALWAYS_RESTART, simpleLoop, 3)
	id, _ := sup.NewChild(ALWAYS_RESTART, simpleLoop, 3)

	time.Sleep(time.Millisecond)

	sup.StopChild(id)

	time.Sleep(time.Millisecond * 100)

	_, _, stopped, _ := sup.Stats()

	if stopped != 1 {
		t.Error("stop supervisor failed")
	}

}

func TestSupMultiWorkers(t *testing.T) {
	sup := NewSupervisor()

	for i := 0; i < 100; i++ {
		sup.NewChild(ALWAYS_RESTART, simpleLoopWithPanic, i)
	}

	for i := 0; i < 100; i++ {
		sup.NewChild(ERROR_RESTART, simpleLoop, i)
	}

	time.Sleep(2 * time.Second)

	total, running, stopped, restarting := sup.Stats()

	sup.Stop()

	fmt.Printf("Total: %d, Running: %d, Stopped: %d, Restarting: %d\n", total, running, stopped, restarting)

	if stopped != 100 || restarting+running != 100 {
		t.Error("has children status failed")
	}
}

func TestSupContext(t *testing.T) {
	sup := NewSupervisorWithContext(context.Background())

	sup.NewChild(NO_RESTART, simpleLoopWithContext, 3)

	time.Sleep(time.Millisecond * 100)

	sup.Stop()
}
