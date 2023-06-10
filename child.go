package easyworker

import (
	"context"
	"fmt"
	"log"
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
	iCHILD_TASK_DONE

	// Child is running task.
	RUNNING = iota

	// Child is restarting.
	RESTARTING

	// Child is stopped.
	STOPPED

	// Child is force to quit. In this state child wait for finish task before quit.
	FORCE_QUIT

	// Child is standby
	STANDBY
)

var (
	// use to store last id of child. id is auto_increment.
	lastChildId atomic.Int64
)

func getNewChildId() int64 {
	return lastChildId.Add(1)
}

/*
A child struct that hold information a bout task, restart strategy.
*/
type Child struct {
	id           int64
	restart_type int
	cmdCh        chan msg

	state     atomic.Int64
	restarted atomic.Int64
	failed    atomic.Int64

	fun    any
	params []any
	ctx    context.Context

	result []any
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

	childId := getNewChildId()

	return Child{
		id:           childId,
		restart_type: restart,
		fun:          fun,
		params:       params,
	}, nil
}

/*
Get child's id.
*/
func (c *Child) Id() int64 {
	return c.id
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
			id:      int(c.id),
			msgType: iCHILD_TASK_DONE,
		}

		// catch if panic by child code.
		if r := recover(); r != nil {
			log.Println(c.id, ", worker was panic, ", r)
			msg.msgType = iCHILD_PANIC
			msg.data = r
			c.incFailed()
		} else {
			c.updateState(STOPPED)
		}

		c.cmdCh <- msg
	}()

	var err error

	c.updateState(RUNNING)

l:
	for {
		// call user define function.
		c.result, err = invokeFun(c.fun, c.params...)

		if err != nil {
			c.incFailed()
			log.Println(c.id, "call user function failed, reason:", err)
		}

		switch c.restart_type {
		case ALWAYS_RESTART:
			if c.getState() == FORCE_QUIT {
				break l
			}
		case ERROR_RESTART:
			if err == nil || c.getState() == FORCE_QUIT {
				log.Println(c.id, "done, child no re-run")
				break l
			}
		case NO_RESTART:
			// always exit.
			break l
		}

		c.incRestarted()
	}
}

func (c *Child) stop() {
	log.Println(c.id, "force stop")
	c.updateState(FORCE_QUIT)
}

func (c *Child) updateState(newStatus int) {
	c.state.Store(int64(newStatus))
}

func (c *Child) getState() int64 {
	return c.state.Load()
}

func (c *Child) canRun() bool {

	return int(c.state.Load()) != FORCE_QUIT
}

func (c *Child) incRestarted() {
	c.restarted.Store(c.restarted.Add(1))
}

func (c *Child) incFailed() {
	c.failed.Store(c.failed.Add(1))
}

func (c *Child) getRestarted() int64 {
	return c.restarted.Load()
}

func (c *Child) getFailed() int64 {
	return c.failed.Load()
}

/*
Return current status & statistic of Child.
*/
func (c *Child) GetStats() (status int64, restarted int64, failed int64) {
	status = c.getState()
	restarted = c.getRestarted()
	failed = c.getFailed()

	return
}

/*
Get result from last run.
Result is slice of any.
Length of slice is number of parameter return from user function.
Cast type to get right value.
*/
func (c *Child) GetResult() []any {
	return c.result
}
