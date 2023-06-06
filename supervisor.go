package easyworker

import (
	"fmt"
	"log"
)

var (
	// use to store last id of supervisor. id is auto_increment.
	supervisorLastId int
)

/*
A supervisor that controll children(workers).
Supervisor privides interface to user with simple APIs.
You can run multi instance supervisor in your application.
*/
type Supervisor struct {
	id       int
	children map[int]*Child
	cmdCh    chan msg
}

/*
Create new supervisor.
*/
func NewSupervisor() Supervisor {
	supervisorLastId++

	ret := Supervisor{
		id:       supervisorLastId,
		children: make(map[int]*Child),
		cmdCh:    make(chan msg),
	}

	ret.start()
	return ret
}

/*
Add directly child to a supervisor.
*/
func (s *Supervisor) NewChild(restart int, fun any, params ...any) (id int, err error) {
	if restart < ALWAYS_RESTART || restart > NO_RESTART {
		err = fmt.Errorf("in correct restart type, input: %d", restart)
		return
	}

	if err = verifyFunc(fun); err != nil {
		return
	}

	childLastId++

	child := &Child{
		id:           childLastId,
		restart_type: restart,
		fun:          fun,
		params:       params,
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

	child.run()
}

/*
Get child from id.
Return nil if id isn't existed.
*/
func (s *Supervisor) GetChild(id int) *Child {
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
				child = s.children[event.id]
				if child.canRun() && (child.restart_type == ALWAYS_RESTART || child.restart_type == ERROR_RESTART) {
					child.updateStatus(RESTARTING)
					log.Println("restarting child:", child.id)
					child.run()
				} else {
					log.Println("child:", child.id, "stopped")
					child.updateStatus(STOPPED)
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
func (s *Supervisor) StopChild(id int) {
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
