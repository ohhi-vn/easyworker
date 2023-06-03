package easyworker

import "fmt"

var (
	// use to store last id of supervisor. id is auto_increment.
	supervisorLastId int
)

type cmd struct {
	id      int
	typeCmd int
	data    interface{}
}

type Supervisor struct {
	id       int
	children map[int]*Child
	cmdCh    chan cmd
}

func NewSupervisor() Supervisor {
	supervisorLastId++

	ret := Supervisor{
		id:       supervisorLastId,
		children: make(map[int]*Child),
		cmdCh:    make(chan cmd),
	}

	ret.start()
	return ret
}

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

	go child.run()

	return
}

func (s *Supervisor) AddChild(child *Child) {
	s.children[child.id] = child
	child.cmdCh = s.cmdCh

	go child.run()
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
			switch event.typeCmd {
			case iCHILD_PANIC:
				child = s.children[event.id]
				if child.canRun() && (child.restart_type == ALWAYS_RESTART || child.restart_type == NORMAL_RESTART) {
					child.updateStatus(iCHILD_RESTARTING)
					fmt.Println("restarting child:", child.id)
					go child.run()
				} else {
					fmt.Println("child:", child.id, "stopped")
					child.updateStatus(iCHILD_STOPPED)
				}

			}
		}
	}()
}

func (s *Supervisor) Stop() {
	for _, child := range s.children {
		child.stop()
	}
}
