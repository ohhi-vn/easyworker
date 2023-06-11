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

	// call user function failed/panic
	SIGNAL_FAILED
)

const ()

// Signal was sent when Goroutine is no longer run (done/failed/panic).
type GoSignal struct {
	// Monitor reference id.
	RefId int64

	// Kind of signal (SIGNAL_DONE, SIGNAL_FAILED)
	Signal int
}

// channel for send signal.
type monitorChan chan GoSignal

// A struct wrap goroutine to handle panic and can re-run easily.
type Go struct {
	lock sync.Mutex

	state          atomic.Int64
	id             int64
	panicListeners map[int64]monitorChan

	fun    any
	params []any

	// store result
	result []any
}

var (
	lastRefId atomic.Int64
)

func getNewRefId() int64 {
	return lastRefId.Add(1)
}

/*
Create new Go struct.
The first parameter is user function.
The second parameter and more is parameter for parameter of user function.
*/
func NewGo(fun any, params ...any) (ret *Go, retErr error) {
	if retErr = verifyFunc(fun); retErr != nil {
		return
	}

	id := getNewRefId()

	ret = &Go{
		id:             id,
		fun:            fun,
		params:         params,
		panicListeners: make(map[int64]monitorChan),
	}
	ret.state.Store(STANDBY)

	return
}

/*
Create Go and run.
Same with function NewGo but run after create Go struct.
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

/*
Run and wait for task done.
Return true if task done in normally. False for failed. Error is cannot run task.
*/
func (g *Go) RunAndWait() (bool, error) {
	_, ch := g.Monitor()

	err := g.Run()
	if err != nil {
		return false, err
	}

	msg := <-ch

	return msg.Signal == SIGNAL_DONE, nil
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

	g.result = nil
	g.state.Store(STOPPED)
}

/*
Start Go to process task.
The function can call many.
*/
func (g *Go) Run() error {
	if g.state.Load() == STOPPED {
		return errors.New("Go cannot run, it stopped")
	}

	go g.run_task()
	return nil
}

func (g *Go) run_task() {
	g.state.Store(RUNNING)
	msg := GoSignal{}
	defer func() {
		// catch if panic by child code.
		if r := recover(); r != nil {
			msg.Signal = SIGNAL_FAILED
			if printLog {
				log.Println(g.id, ", Go was panic, ", r)
			}
		}

		g.pushSignal(msg)
		g.state.Store(STANDBY)
	}()

	var (
		err    error
		result []any
	)

	//log.Println("Go run, params:", g.params)

	// call user define function.
	result, err = invokeFun(g.fun, g.params...)

	g.lock.Lock()
	g.result = result
	g.lock.Unlock()

	if err != nil {
		msg.Signal = SIGNAL_FAILED
		if printLog {
			log.Println(g.id, "Go call user function failed, reason:", err)
		}
	}
}

/*
Return state of Go.
Kind of state:
  - RUNNING: Task is running.
  - STOPPED: Stop by user. In this state user cannot run again.
  - STANDBY: Task is standby wait for start or just done task.
*/
func (g *Go) State() int64 {
	return g.state.Load()
}

/*
Get result from last run.
Result is slice of any.
Length of slice is number of parameter return from user function.
Cast to right type for value.
*/
func (g *Go) GetResult() []any {
	g.lock.Lock()
	defer g.lock.Unlock()

	return g.result
}
