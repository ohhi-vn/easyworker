package easyworker

import (
	"testing"
	"time"
)

func loopRun2(a int) {
	for i := 0; i < a; i++ {
		time.Sleep(time.Millisecond)
	}
}

func TestChildIncorrectedParams(t *testing.T) {
	_, err := NewChild(ALWAYS_RESTART, "Loop2", 5)

	if err == nil {
		t.Error("missed checking function from user")
	}

	_, err = NewChild(1000, loopRun2, 5)

	if err == nil {
		t.Error("missed checking restart strategy from user")
	}
}
