package easyworker

import (
	"errors"
	"log"
	"sync"
	"sync/atomic"
)

const (
	// call user function success.
	SIGNAL_DONE = iota

	// call user function failed
	SIGNAL_FAILED
)

// Signal was sent when Goroutine is no longer run (done/failed/panic).
type GoSignal struct {
	// Monitor reference id.
	RefId int64
	// Kind of signal (SIGNAL_DONE, SIGNAL_FAILED)
	Signal int
}

type monitorChan chan GoSignal

// A goroutine can be monitored.
type Go struct {
	lock sync.Mutex

	stopped        bool
	id             int64
	panicListeners map[int64]monitorChan

	fun    any
	params []any
}

var (
	lastRefId atomic.Int64
)

func getNewRefId() int64 {
	return lastRefId.Add(1)
}

/*
Create new Go.
*/
func NewGo(fun any, params ...any) (ret *Go, retErr error) {
	if retErr = verifyFunc(fun); retErr != nil {
		return
	}

	id := getNewRefId()

	return &Go{
		id:             id,
		fun:            fun,
		params:         params,
		panicListeners: make(map[int64]monitorChan),
	}, nil
}

/*
Create Go and run.
*/
func NewGoAndRun(fun any, params ...any) (ret *Go, retErr error) {
	ret, retErr = NewGo(fun, params...)
	if retErr != nil {
		return
	}

	retErr = ret.Run()

	return
}

/*
Used for receiving a signal when Go done or failed.
Function return unique reference id and a channel for receiving signal.
*/
func (g *Go) Monitor() (int64, <-chan GoSignal) {
	refId := getNewRefId()
	// always set buffer to 1 for async and run_task can exit.
	ch := make(monitorChan, 1)

	g.lock.Lock()
	defer g.lock.Unlock()

	g.panicListeners[refId] = ch

	return refId, ch
}

/*
Remove a monitor reference.
After demonitor channel will be closed.
*/
func (g *Go) Demonitor(refId int64) {
	g.lock.Lock()
	defer g.lock.Unlock()

	ch, existed := g.panicListeners[refId]
	if existed {
		close(ch)
		delete(g.panicListeners, refId)
	}
}

func (g *Go) pushSignal(msg GoSignal) {
	g.lock.Lock()
	defer g.lock.Unlock()

	for refId, ch := range g.panicListeners {
		msg.RefId = refId
		ch <- msg
	}
}

/*
Stop Go just for clean data in internal struct.
Call Stop after Go process task done.
*/
func (g *Go) Stop() {
	g.lock.Lock()
	defer g.lock.Unlock()

	for refId, ch := range g.panicListeners {
		close(ch)
		delete(g.panicListeners, refId)
	}

	g.stopped = true
}

/*
Start Go to process task.
The function can call many.
*/
func (g *Go) Run() error {
	if g.stopped {
		return errors.New("Go cannot run, it stopped")
	}

	go g.run_task()
	return nil
}

func (g *Go) run_task() {
	msg := GoSignal{}
	defer func() {
		// catch if panic by child code.
		if r := recover(); r != nil {
			msg.Signal = SIGNAL_FAILED
			log.Println(g.id, ", goroutine was panic, ", r)
		}

		g.pushSignal(msg)
	}()

	var err error

	//log.Println("Go run, params:", g.params)

	// call user define function.
	_, err = invokeFun(g.fun, g.params...)

	if err != nil {
		msg.Signal = SIGNAL_FAILED
		log.Println(g.id, "goroutine call user function failed, reason:", err)
	}
}
