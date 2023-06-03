package easyworker

import (
	"fmt"
	"reflect"
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
	CHILD_PANIC = iota
	CHILD_DONE
	CHILD_RUNING
	CHILD_RESTARTING
	CHILD_STOPPED
)

var (
	// use to store last id of child. id is auto_increment.
	childLastId int
)

type Child struct {
	id           int
	restart_type int
	cmdCh        chan cmd
	status       int

	fun    interface{}
	params []interface{}
}

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
			typeCmd: CHILD_DONE,
		}
		if r := recover(); r != nil {
			fmt.Println(c.id, ", worker was panic, ", r)
			msg.typeCmd = CHILD_PANIC
			msg.data = r
		}

		c.cmdCh <- msg
	}()

	var (
		ret reflect.Value
		err error
	)

	c.status = CHILD_RUNING
l:
	for {
		// call user define function.
		ret, err = invokeFun(c.fun, c.params...)

		switch c.restart_type {
		case ALWAYS_RESTART:
			if err != nil {
				fmt.Println(c.id, "failed, reason:", err)

			}
			fmt.Println(c.id, "child re-run")
			continue l
		case NORMAL_RESTART:
			if err == nil {
				fmt.Println(c.id, "done, child no re-run")
				break l
			} else {
				fmt.Println(c.id, "failed, child re-run, reason:", err)
				continue l
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
