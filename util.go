package easyworker

import (
	"runtime"
)

const (
	iDEFAULT_RETRY = 3
)

/*
Return default number of workers.
Default number of workers is equal number cores of CPU.
*/
func DefaultNumWorkers() int {
	return runtime.NumCPU()
}

/*
Return default retry times of package.
Default is 3 times if workers failed.
*/
func DefaultRetryTimes() int {
	return iDEFAULT_RETRY
}
