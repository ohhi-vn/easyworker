package easyworker

import (
	"fmt"
	"log"
	"reflect"
	"sync/atomic"
)

const (
	// Always restart child for both case, done task normally or panic.
	ALWAYS_RESTART = iota

	// Just restart if child got an panic.
	ERROR_RESTART

	// No restart child for any reason.
	NO_RESTART
)

const (
	iCHILD_PANIC = iota
	iCHILD_DONE
	iCHILD_RUNNING
	iCHILD_RESTARTING
	iCHILD_STOPPED
	iCHILD_FORCE_QUIT
)

var (
	// use to store last id of child. id is auto_increment.
	childLastId int
)

/*
A child struct that hold information a bout task, restart strategy.
*/
type Child struct {
	id           int
	restart_type int
	cmdCh        chan msg
	status       atomic.Int64

	fun    any
	params []any
}

/*
Create new child.
*/
func NewChild(restart int, fun any, params ...any) (ret Child, retErr error) {
	if restart < ALWAYS_RESTART || restart > NO_RESTART {
		retErr = fmt.Errorf("in correct restart type, input: %d", restart)
		return
	}

	if retErr = verifyFunc(fun); retErr != nil {
		return
	}

	childLastId++

	return Child{
		id:           childLastId,
		restart_type: restart,
		fun:          fun,
		params:       params,
	}, nil
}

/*
Start goroutine to execute task.
*/
func (c *Child) run() {
	go c.run_task()
}

/*
Run task.
*/
func (c *Child) run_task() {
	defer func() {
		msg := msg{
			id:      c.id,
			msgType: iCHILD_DONE,
		}
		if r := recover(); r != nil {
			log.Println(c.id, ", worker was panic, ", r)
			msg.msgType = iCHILD_PANIC
			msg.data = r
		} else {
			c.updateStatus(iCHILD_STOPPED)
		}

		c.cmdCh <- msg
	}()

	var (
		ret reflect.Value
		err error
	)

	c.updateStatus(iCHILD_RUNNING)

l:
	for {
		if c.getStatus() == iCHILD_FORCE_QUIT {
			log.Println(c.id, "force stop")
			break l
		}

		// call user define function.
		ret, err = invokeFun(c.fun, c.params...)

		switch c.restart_type {
		case ALWAYS_RESTART:
			if err != nil {
				log.Println(c.id, "failed, reason:", err)
			}
			log.Println(c.id, "child re-run")
		case ERROR_RESTART:
			if err == nil {
				log.Println(c.id, "done, child no re-run")
				break l
			} else {
				log.Println(c.id, "failed, child re-run, reason:", err)
			}
		case NO_RESTART:
			if err != nil {
				log.Println(c.id, "failed, no re-run, reason:", err)
				break l
			}
		}

	}

	if err != nil {
		log.Println(c.id, ", call function failed, error: ", err)
	} else {
		log.Println(c.id, ", function return ", ret)
	}
}

func (c *Child) stop() {
	log.Println(c.id, "force stop...")
	c.updateStatus(iCHILD_FORCE_QUIT)
}

func (c *Child) updateStatus(newStatus int) {
	c.status.Store(int64(newStatus))
	log.Println(c.id, "status after set", c.getStatus())
}

func (c *Child) getStatus() int {
	return int(c.status.Load())
}

func (c *Child) canRun() bool {

	return int(c.status.Load()) != iCHILD_FORCE_QUIT
}
