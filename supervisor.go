package easyworker

import "fmt"

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
func (s *Supervisor) NewChild(restart int, fun interface{}, params ...interface{}) (retErr error) {
	if restart < ALWAYS_RESTART || restart > NO_RESTART {
		retErr = fmt.Errorf("in correct restart type, input: %d", restart)
		return
	}

	if retErr = verifyFunc(fun); retErr != nil {
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
make a goroutine to handle event from children.
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
					child.updateStatus(iCHILD_RESTARTING)
					fmt.Println("restarting child:", child.id)
					child.run()
				} else {
					fmt.Println("child:", child.id, "stopped")
					child.updateStatus(iCHILD_STOPPED)
				}

			}
		}
	}()
}

/*
Supervisor will send stop signal to children.
Child after process your function will check the signal and stop.
*/
func (s *Supervisor) Stop() {
	for _, child := range s.children {
		child.stop()
	}
}
