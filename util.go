package easyworker

import (
	"runtime"
)

const (
	iDEFAULT_RETRY = 3
)

func DefaultNumWorkers() int {
	return runtime.NumCPU()
}

func DefaultRetryTimes() int {
	return iDEFAULT_RETRY
}
