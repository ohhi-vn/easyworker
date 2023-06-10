package easyworker

import (
	"testing"
	"time"
)

func loopRun2(a int) (ret int) {
	for i := 0; i < a; i++ {
		ret += i
		time.Sleep(time.Millisecond)
	}
	return
}

func TestChildDefault(t *testing.T) {
	child, err := NewChild(ALWAYS_RESTART, loopRun2, 5)

	if err != nil {
		t.Error("create child failed, ", err)
		return
	}

	if child.Id() != lastChildId.Load() {
		t.Error("wrong id for child")
		return
	}

	state, started, failed := child.GetStats()

	if state != int64(STANDBY) || failed != 0 || started != 0 {
		t.Error("incorrect default value, ", state, started, failed)
	}

	if !child.canRun() {
		t.Error("incorrect state")
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
