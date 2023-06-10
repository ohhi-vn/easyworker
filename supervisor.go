package easyworker

import (
	"context"
	"fmt"
	"log"
)

type key int

const (
	CTX_SUP_ID key = iota
	CTX_CHILD_ID
)

/*
A supervisor that controll children(workers).
Supervisor privides interface to user with simple APIs.
You can run multi instance supervisor in your application.
*/
type Supervisor struct {
	id       int64
	children map[int64]*Child
	cmdCh    chan msg
	ctx      context.Context
}

/*
Create new supervisor.
*/
func NewSupervisor() (ret Supervisor) {
	newId := getNewSupId()

	ret = Supervisor{
		id:       newId,
		children: make(map[int64]*Child),
		cmdCh:    make(chan msg),
	}

	listSup.add(&ret)

	ret.start()

	return
}

/*
Create new supervisor with context.
*/
func NewSupervisorWithContext(ctx context.Context) (ret Supervisor) {
	if ctx == nil {
		panic("context for supervisor is nil")
	}

	ret = NewSupervisor()
	ret.ctx = context.WithValue(ctx, CTX_SUP_ID, ret.id)
	listSup.add(&ret)

	ret.start()

	return
}

/*
Add directly child to a supervisor.
*/
func (s *Supervisor) NewChild(restart int, fun any, params ...any) (id int64, err error) {
	if restart < ALWAYS_RESTART || restart > NO_RESTART {
		err = fmt.Errorf("in correct restart type, input: %d", restart)
		return
	}

	if err = verifyFunc(fun); err != nil {
		return
	}

	childId := getNewChildId()

	var (
		paramsWithCtx []any
		ctx           context.Context
	)

	// add context to first param of task
	if s.ctx != nil {
		ctx = context.WithValue(s.ctx, CTX_CHILD_ID, childId)
		paramsWithCtx = make([]any, len(params)+1)
		paramsWithCtx[0] = ctx
		copy(paramsWithCtx[1:], params)
	} else {
		paramsWithCtx = params
		ctx = nil
	}

	child := &Child{
		id:           childId,
		restart_type: restart,
		fun:          fun,
		params:       paramsWithCtx,
		ctx:          ctx,
	}

	s.children[child.id] = child
	child.cmdCh = s.cmdCh

	child.run()

	id = child.id
	return
}

/*
Add existed child to supervisor.
A child can add to run in one or more supervisor.
*/
func (s *Supervisor) AddChild(child *Child) {
	s.children[child.id] = child
	child.cmdCh = s.cmdCh

	if s.ctx != nil {
		// add context to first param of task
		ctx := context.WithValue(s.ctx, CTX_CHILD_ID, child.id)
		paramsWithCtx := make([]any, len(child.params)+1)
		paramsWithCtx[0] = ctx
		copy(paramsWithCtx[1:], child.params)
		child.params = paramsWithCtx
	}

	child.run()
}

/*
Get child from id.
Return nil if id isn't existed.
*/
func (s *Supervisor) GetChild(id int64) *Child {
	child := s.children[id]
	return child
}

/*
Make a goroutine to handle event from children. Restart children if needed.
*/
func (s *Supervisor) start() {
	go func() {
		var (
			child *Child
		)
		for {
			event := <-s.cmdCh
			switch event.msgType {
			case iCHILD_PANIC:
				child = s.children[int64(event.id)]
				if child.canRun() && (child.restart_type == ALWAYS_RESTART || child.restart_type == ERROR_RESTART) {
					child.updateState(RESTARTING)
					log.Println("restarting child:", child.id)
					child.run()
				} else {
					log.Println("child:", child.id, "stopped")
					child.updateState(STOPPED)
				}

			}
		}
	}()
}

/*
Supervisor will send stop signal to children.
Children after process your function will check the signal and stop.
In this case, ALWAYS_RESTART & ERROR_RESTART will be ignored.
*/
func (s *Supervisor) Stop() {
	for _, child := range s.children {
		child.stop()
	}
}

/*
Supervisor will send stop signal to a child.
Child after process your function will check the signal and stop.
In this case, ALWAYS_RESTART & ERROR_RESTART will be ignored.
*/
func (s *Supervisor) StopChild(id int64) {
	if child, existed := s.children[id]; existed {
		child.stop()
	}
}

/*
Return statistics of supervisor.
total: Number of children in supervisor.
running: Number of children are running.
stopped: Number of children are stopped.
restarting: Number of children are restarting.
*/
func (s *Supervisor) Stats() (total, running, stopped, restarting int) {
	total = len(s.children)
	for _, child := range s.children {
		switch child.getState() {
		case RUNNING:
			running++
		case RESTARTING:
			restarting++
		default:
			stopped++
		}
	}

	return
}
