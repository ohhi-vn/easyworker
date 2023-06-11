package easyworker

import (
	"log"
	"time"
)

const (
	iSTREAM = -1

	iERROR = iota
	iFATAL_ERROR
	iSUCCESS
	iRETRY
	iCANCEL
	iRUN
	iQUIT
	iTASK
)

// struct carry task/command for worker.
type msg struct {
	id      int
	msgType int
	data    any
}

// worker's information.
type worker struct {
	// worker's id
	id int64

	// retry time, define by user.
	retryTimes int

	// sleep time between re-run.
	retrySleep int

	// function, define by user.
	fun any

	// command channel, supervisor uses to send command to worker.
	cmd chan msg

	// input channel, worker receives task (params) then run with fun.
	inputCh chan msg

	// output channel, worker send back result to supervisor.
	resultCh chan msg
}

/*
start worker with options in struct.
after start worker will wait task from supervisor.
after task done, worker will send result back to supervisor with id of task.
*/
func (w *worker) run() {
	defer func() {
		if r := recover(); r != nil {
			if printLog {
				log.Println(w.id, ", worker was panic, ", r)
			}
			w.resultCh <- msg{id: int(w.id), msgType: iFATAL_ERROR, data: r}
		}
	}()

	var (
		task msg
		ret  []any
		err  error
	)

	for {
		select {
		case task = <-w.inputCh:
		case cmd := <-w.cmd:
			// receive a quit signal.
			if cmd.msgType == iQUIT {
				if printLog {
					log.Println(w.id, "is exited")
				}
				return
			}
		}

		switch task.msgType {
		case iTASK:
			args := task.data.([]any)

			for i := 0; i <= w.retryTimes; i++ {
				if i > 0 {
					time.Sleep(time.Millisecond * time.Duration(w.retrySleep))
					if printLog {
						log.Println(w.id, ", retry(", i, ") function with last args")
					}
				}
				ret, err = invokeFun(w.fun, args...)
				if err == nil {
					break
				}
			}

			if err != nil {
				if printLog {
					log.Println(w.id, ", call function failed, error: ", err)
				}
				w.resultCh <- msg{id: task.id, msgType: iERROR, data: err}
			} else {
				w.resultCh <- msg{id: task.id, msgType: iSUCCESS, data: ret}
			}
		}
	}
}
