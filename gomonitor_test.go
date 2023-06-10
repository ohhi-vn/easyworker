package easyworker

import (
	"fmt"
	"testing"
	"time"
)

func TestGoNotAFunc(t *testing.T) {
	_, err := NewGo("hello")

	if err == nil {
		t.Error("cannot cat incorrect func, ", err)
	}
}

func TestGoNotAFunc2(t *testing.T) {
	_, err := NewGoAndRun("hello")

	if err == nil {
		t.Error("cannot cat incorrect func, ", err)
	}
}

func TestGoRunNoArg(t *testing.T) {
	g, err := NewGo(simpleLoopNoArg)

	if err != nil {
		t.Error("create go failed, ", err)
		return
	}

	refId, ch := g.Monitor()

	g.Run()

	select {
	case sig := <-ch:
		if refId != sig.RefId || SIGNAL_DONE != sig.Signal {
			t.Error("return signal incorrect")
		}
	case <-time.After(time.Second):
		t.Error("timed out")
	}
}

func TestGoRun(t *testing.T) {
	g, err := NewGo(loopRun2, 5)

	if err != nil {
		t.Error("create go failed, ", err)
		return
	}

	refId, ch := g.Monitor()

	g.Run()

	select {
	case sig := <-ch:
		if refId != sig.RefId || SIGNAL_DONE != sig.Signal {
			t.Error("return signal incorrect")
		}
	case <-time.After(time.Second):
		t.Error("timed out")
	}
}

func TestGoRun2(t *testing.T) {
	g, err := NewGo(simpleLoopWithPanic, 5)

	if err != nil {
		t.Error("create go failed, ", err)
		return
	}

	refId, ch := g.Monitor()

	g.Run()

	select {
	case sig := <-ch:
		if refId != sig.RefId || SIGNAL_FAILED != sig.Signal {
			t.Error("return signal incorrect")
		}
	case <-time.After(time.Second):
		t.Error("timed out")
	}
}

func TestGoRun3(t *testing.T) {
	g, err := NewGo(simpleLoopWithPanic, 5)

	if err != nil {
		t.Error("create go failed, ", err)
		return
	}

	refId, ch := g.Monitor()

	g.Run()

	for i := 0; i < 3; i++ {
		select {
		case sig := <-ch:
			if refId != sig.RefId || SIGNAL_FAILED != sig.Signal {
				t.Error("return signal incorrect", sig)
				return
			}

			fmt.Println("re-run Go, times: ", i)
			g.Run()
		case <-time.After(time.Second):
			t.Error("timed out")
			return
		}
	}
}

func TestGoNewAndRun(t *testing.T) {
	g, err := NewGoAndRun(loopRun2, 5)

	if err != nil {
		t.Error("create go failed, ", err)
		return
	}

	time.Sleep(time.Second)

	r := g.GetResult()
	if r[0].(int) != 10 {
		t.Error("incorrect result")
	}
}

func TestGoRunAndWait(t *testing.T) {
	g, err := NewGo(loopRun2, 5)

	if err != nil {
		t.Error("create go failed, ", err)
		return
	}

	g.RunAndWait()

	r := g.GetResult()
	if r[0].(int) != 10 {
		t.Error("incorrect result")
	}
}

func TestGoDemonitor(t *testing.T) {
	g, err := NewGo(simpleLoopWithPanic, 5)

	if err != nil {
		t.Error("create go failed, ", err)
		return
	}

	refId, ch := g.Monitor()

	g.Run()

	for i := 0; i < 2; i++ {
		select {
		case sig, more := <-ch:

			// channel is closed
			if !more && sig.RefId == 0 {
				return
			}
			if refId != sig.RefId || SIGNAL_FAILED != sig.Signal {
				t.Error("return signal incorrect", sig)
				return
			}

			fmt.Println("re-run Go, times: ", i)
			g.Demonitor(refId)

			g.Run()
		case <-time.After(time.Second):
			t.Error("timed out")
			return
		}
	}
}

func TestGoMultiMonitor(t *testing.T) {
	g, err := NewGo(loopRun2, 5)

	if err != nil {
		t.Error("create go failed, ", err)
		return
	}

	okCh := make(chan bool)

	go func(out chan<- bool) {
		refId, ch := g.Monitor()

		sig := <-ch
		if refId != sig.RefId || SIGNAL_DONE != sig.Signal {
			out <- false
		} else {
			fmt.Println("ref:", refId, "ok")
			out <- true
		}
	}(okCh)

	go func(out chan<- bool) {
		refId, ch := g.Monitor()

		sig := <-ch
		if refId != sig.RefId || SIGNAL_DONE != sig.Signal {
			out <- false
		} else {
			fmt.Println("ref:", refId, "ok")
			out <- true
		}
	}(okCh)

	g.Run()

	for i := 0; i < 2; i++ {
		select {
		case ret := <-okCh:
			if !ret {
				t.Error("return signal incorrect")
				return
			}
		case <-time.After(time.Second):
			t.Error("timed out")
			return
		}
	}
}

func TestGoMassMonitor(t *testing.T) {
	g, err := NewGo(loopRun2, 100)
	num := 10000

	if err != nil {
		t.Error("create go failed, ", err)
		return
	}

	okCh := make(chan bool)

	for i := 0; i < num; i++ {
		go func(out chan<- bool) {
			refId, ch := g.Monitor()

			sig := <-ch
			if refId != sig.RefId || SIGNAL_DONE != sig.Signal {
				out <- false
			} else {
				out <- true
			}
		}(okCh)
	}

	g.Run()

	for i := 0; i < num; i++ {
		select {
		case ret := <-okCh:
			if !ret {
				t.Error("return signal incorrect")
				return
			}
		case <-time.After(5 * time.Second):
			t.Error("timed out")
			return
		}
	}
}

func TestGoMassMonitor2(t *testing.T) {
	num := 10000

	okCh := make(chan bool, num)

	for i := 0; i < num; i++ {
		g, err := NewGo(loopRun2, 100)
		if err != nil {
			t.Error("create go failed, ", err)
			return
		}
		go func(out chan<- bool) {
			refId, ch := g.Monitor()

			sig := <-ch
			if refId != sig.RefId || SIGNAL_DONE != sig.Signal {
				out <- false
			} else {
				out <- true
			}
		}(okCh)

		g.Run()
	}

	for i := 0; i < num; i++ {
		select {
		case ret := <-okCh:
			if !ret {
				t.Error("return signal incorrect")
				return
			}
		case <-time.After(5 * time.Second):
			t.Error("timed out")
			return
		}
	}
}

func TestGoStop(t *testing.T) {
	g, err := NewGo(loopRun2, 5)

	if err != nil {
		t.Error("create go failed, ", err)
		return
	}

	refId, ch := g.Monitor()

	g.Run()

	select {
	case sig := <-ch:
		if refId != sig.RefId || SIGNAL_DONE != sig.Signal {
			t.Error("return signal incorrect")
			return
		}
	case <-time.After(time.Second):
		t.Error("timed out")
		return
	}

	g.Stop()

	if g.GetResult() != nil {
		t.Error("not reset result")
		return
	}

	ok, _ := g.RunAndWait()
	if ok {
		t.Error("expect to cannot re-run ")
	}
}

func TestGoRerun(t *testing.T) {
	g, err := NewGo(loopRun2, 5)

	if err != nil {
		t.Error("create go failed, ", err)
		return
	}

	refId, ch := g.Monitor()

	g.Run()

	select {
	case sig := <-ch:
		if refId != sig.RefId || SIGNAL_DONE != sig.Signal {
			t.Error("return signal incorrect")
			return
		}
	case <-time.After(time.Second):
		t.Error("timed out")
		return
	}

	ok, err := g.RunAndWait()
	if !ok {
		t.Error("expect to can re-run ", err)
	} else {
		if g.GetResult()[0] != 10 {
			t.Error("incorrect result")
		}
	}
}
