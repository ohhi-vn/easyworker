package easyworker

import (
	"fmt"
	"reflect"
	"sync/atomic"
)

const (
	// Always restart child
	ALWAYS_RESTART = iota

	// Just restart if child got an error
	NORMAL_RESTART

	// No restart child
	NO_RESTART
)

const (
	iCHILD_PANIC = iota
	iCHILD_DONE
	iCHILD_RUNING
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
	cmdCh        chan cmd
	status       atomic.Int64

	fun    interface{}
	params []interface{}
}

/*
Create new child.
*/
func NewChild(restart int, fun interface{}, params ...interface{}) (ret Child, retErr error) {
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

func (c *Child) run() {
	defer func() {
		msg := cmd{
			id:      c.id,
			typeCmd: iCHILD_DONE,
		}
		if r := recover(); r != nil {
			fmt.Println(c.id, ", worker was panic, ", r)
			msg.typeCmd = iCHILD_PANIC
			msg.data = r
		}

		c.cmdCh <- msg
	}()

	var (
		ret reflect.Value
		err error
	)

	c.updateStatus(iCHILD_RUNING)

l:
	for {
		fmt.Println(c.id, "status in for", c.getStatus())
		if c.getStatus() == iCHILD_FORCE_QUIT {
			fmt.Println(c.id, "force stop")
			break l
		}

		// call user define function.
		ret, err = invokeFun(c.fun, c.params...)

		switch c.restart_type {
		case ALWAYS_RESTART:
			if err != nil {
				fmt.Println(c.id, "failed, reason:", err)
			}
			fmt.Println(c.id, "child re-run")
		case NORMAL_RESTART:
			if err == nil {
				fmt.Println(c.id, "done, child no re-run")
				break l
			} else {
				fmt.Println(c.id, "failed, child re-run, reason:", err)
			}
		case NO_RESTART:
			if err != nil {
				fmt.Println(c.id, "failed, no re-run, reason:", err)
				break l
			}
		}

	}

	if err != nil {
		fmt.Println(c.id, ", call function failed, error: ", err)
	} else {
		fmt.Println(c.id, ", function return ", ret)
	}
}

func (c *Child) stop() {
	fmt.Println(c.id, "force stop...")
	c.updateStatus(iCHILD_FORCE_QUIT)
}

func (c *Child) updateStatus(newStatus int) {
	c.status.Store(int64(newStatus))
	fmt.Println(c.id, "status after set", c.getStatus())
}

func (c *Child) getStatus() int {
	return int(c.status.Load())
}

func (c *Child) canRun() bool {

	return int(c.status.Load()) != iCHILD_FORCE_QUIT
}
