package easyworker

import (
	"sync"
	"sync/atomic"
)

type supervisorMan struct {
	lock    sync.RWMutex
	listSup map[int64]*Supervisor
}

var (
	lastSupId atomic.Int64

	listSup supervisorMan = supervisorMan{
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
Remove supervisor.
*/
func RemoveSupervisor(sup *Supervisor) {
	listSup.remove(sup.id)
}

/*
Remove supervisor by id.
User need stop before remove.
*/
func RemoveSupervisorById(id int64) {
	listSup.remove(id)
}

func (sm *supervisorMan) add(sup *Supervisor) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	sm.listSup[sup.id] = sup
}

func (sm *supervisorMan) remove(id int64) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	if sup, existed := sm.listSup[id]; existed {
		for k := range sup.children {
			delete(sup.children, k)
		}
	}

	delete(sm.listSup, id)
}

func (sm *supervisorMan) get(id int64) *Supervisor {
	sm.lock.RLock()
	defer sm.lock.RUnlock()

	return sm.listSup[id]
}
