package easyworker

import (
	"fmt"
	"testing"
	"time"
)

func LoopRun(a int, testSupporter chan int) {
	fmt.Println("LoopRun, param:", a)
	testSupporter <- a
	for i := 0; i < a; i++ {
		time.Sleep(time.Millisecond)
		fmt.Println("loop at", i)
	}
	fmt.Println("Loop exit..")
}

func LoopRunWithPanic(a int, testSupporter chan int) {
	testSupporter <- a
	for i := 0; i < a; i++ {
		time.Sleep(time.Millisecond)
		fmt.Println("loop at", i)
		if i == 1 {
			panic("test loop with panic")
		}
	}
	fmt.Println("Loop exit..")
}

func TestSupAlwaysRestart1(t *testing.T) {
	fmt.Println("test TestSupAlwaysRestart1")
	ch := make(chan int)

	sup := NewSupervisor()

	child, _ := NewChild(ALWAYS_RESTART, LoopRun, 5, ch)

	sup.AddChild(child)

	fmt.Println("start waiting signal from worker")
	counter := 0
l:
	for {
		select {
		case param := <-ch:
			fmt.Println("init param:", param)
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

	sup.AddChild(child)

	counter := 0
l:
	for {
		select {
		case param := <-ch:
			fmt.Println("init param:", param)
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

	child, _ := NewChild(NORMAL_RESTART, LoopRun, 5, ch)

	sup.AddChild(child)

	counter := 0
l:
	for {
		select {
		case <-ch:
			counter++
			if counter > 1 {
				t.Error("unexpected, child was restarted in NORMAL_RESTART strategy, fun run sucessful")
			}
		case <-time.After(time.Second):
			break l
		}
	}
}

func TestSupNormalRestart2(t *testing.T) {
	ch := make(chan int)

	sup := NewSupervisor()

	child, _ := NewChild(NORMAL_RESTART, LoopRunWithPanic, 5, ch)

	sup.AddChild(child)

	counter := 0
l:
	for {
		select {
		case param := <-ch:
			fmt.Println("init param:", param)
			counter++
			if counter > 3 {
				break l
			}
		case <-time.After(time.Second):
			t.Error("timed out")
		}
	}
}
