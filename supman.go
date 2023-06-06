package easyworker

import (
	"sync"
	"sync/atomic"
)

type SupervisorMan struct {
	lock    sync.RWMutex
	listSup map[int64]*Supervisor
}

var (
	lastSupId atomic.Int64

	listSup SupervisorMan = SupervisorMan{
		listSup: make(map[int64]*Supervisor),
	}
)

func getNewSupId() int64 {
	return lastSupId.Add(1)
}

/*
Get supervisor from id.
*/
func GetSupervisor(id int64) *Supervisor {
	return listSup.get(id)
}

/*
Remove supervisor by id.
User need stop before remove.
*/
func RemoveSupervisor(id int64) {
	listSup.remove(id)
}

func (sm *SupervisorMan) add(sup *Supervisor) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	sm.listSup[sup.id] = sup
}

func (sm *SupervisorMan) remove(id int64) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	delete(sm.listSup, id)
}

func (sm *SupervisorMan) get(id int64) *Supervisor {
	sm.lock.RLock()
	defer sm.lock.RUnlock()

	return sm.listSup[id]
}
