package easyworker

import (
	"fmt"
	"testing"
	"time"
)

func LoopRun2(a int) {
	for i := 0; i < a; i++ {
		time.Sleep(time.Second)
		fmt.Println("loop at", i)
	}
	fmt.Println("Loop exit..")
}

func TestChildIncorrectedParams(t *testing.T) {
	_, err := NewChild(ALWAYS_RESTART, "LoopRun", 5)

	if err == nil {
		t.Error("missed checking function from user")
	}

	_, err = NewChild(1000, LoopRun, 5)

	if err == nil {
		t.Error("missed checking restart strategy from user")
	}
}
