package easyworker

import (
	"runtime"
	"testing"
)

func TestDefaultNumWorkers(t *testing.T) {
	if runtime.NumCPU() != DefaultNumWorkers() {
		t.Error("incorrect default number of workers")
	}
}

func TestDefaultRetryTimes(t *testing.T) {
	if iDEFAULT_RETRY != DefaultRetryTimes() {
		t.Error("incorrect default retry times")
	}
}
