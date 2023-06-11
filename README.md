[![Go Reference](https://pkg.go.dev/badge/github.com/manhvu/easyworker.svg)](https://pkg.go.dev/github.com/manhvu/easyworker)

# Introduce

A Golang package for supporting worker supervisor model.
The package help developer easy to run parallel tasks.
It's scalable with minimal effort.

easyworker is inspired by Erlang OTP.

# Design

The package has two main part:

* The supervisor: Manage children, give a task to children & collect result.
* The children: Receive tasks, run tasks then send results to supervisor.

## Supervisor

Start workers and monitor workers.
Send task to workers and get result(for task & stream).
Restart(depends restart strategy) children if they're failed to call user function or panic.
Supervisor has one goroutine for send & manage signal to children.

## Child

Run task with user's function and handle error.
If user's function panic worker will check retry config and re-run (depend restart strategy/retry times) if needed.
Each child has a goroutine to run its task.

# Guide

easyworker support 3 type of workers:

* Supervisor, Start a supervisor for managing children. Children can run many type of tasks.
* Task, Add a list of task and run workers. Workers run same type of task.
* Stream, Start workers then push tasks to workers from channel. Workers run same type of task.

Package use standard log package, if you want log to file please set output for log package.

(*required Go 1.19 or later.*)

## Install

Add package to go.mod

```bash
go get github.com/manhvu/easyworker
```

Import package

```go
import "github.com/manhvu/easyworker"
```

## Use Cases

### Supervisor model

This is used for generic worker(child).
Every children has a owner restart strategy.

Currently, child has three type of restart strategy:

* ALWAYS_RESTART, supervisor always restart children if it panic or done task.
* ERROR_RESTART, supervisor will only restart if children was panic.
* NO_RESTART, supervisor will don't restart children for any reason.

Children will be started after they are added to supervisor.
Chid can be added many times to many supervisor but child can control only by the last supervior.

In restart case, children will re-use last parameters (if task don't change it) of task.

Child, doesn't return any value from task.
You need add code to get value from task if you needed.

Supervisor -> Child -> call user functions

Basic supervisor's flow:

```mermaid
graph LR
User(User code) -->|init supervisor & children|Sup(Supervisor 1)
    Sup-->|add child & run|Child1(Child 1 - Task 1)
    Sup-->|add child & run|Child2(Child 2 - Task 2)
    Sup-->|add child & run|Childn(Child N - Task N)
```

(install extension support mermaid to view flow)

supervisor example:

```go
// example function need to run in child.
loop := func(a int) {
 for i := 0; i < a; i++ {
  time.Sleep(time.Second)
  fmt.Println("loop at", i)
 }
 fmt.Println("loop exit...")
}

// example function run in child. It will panic if counter > 3.
loopWithPanic := func(a int, panicString string) {
 for i := 0; i < a; i++ {
  time.Sleep(time.Second)
  fmt.Println("loop at", i)
  if i > 3 {
   panic(panicString)
  }
 }
    // maybe you won't see this.
 fmt.Println("loopWithPanic exit...")
}

// create a supervisor.
sup := easyworker.NewSupervisor()

// add direct child to supervisor.
sup.NewChild(easyworker.ERROR_RESTART, loop, 5)
sup.NewChild(easyworker.NO_RESTART, func() {
  fmt.Println("Hello")
})

// create a child.
child, _ := easyworker.NewChild(easyworker.ALWAYS_RESTART, loopWithPanic, 5, "test panic")

// add exists child.
sup.AddChild(child)

// or do something you want.
time.Sleep(15 * time.Second)

// stop all worker.
// this function depends how long fun return.
sup.Stop()
```

Supervisor support context by create supervisor by function `NewSupervisorWithContext`.
In case supervisor with context, the first parameter of user function will be context.
Context will include supervisor's id and child's id.

User code can get with key:

* CTX_SUP_ID for supervisor's id
* CTX_CHILD_ID for child's id

Call `GetSupervisor` to get supervisor from supervisor's id.

To get child please call `GetChild` method of supervisor.

Example:

```go
// basic func with context.
loopWithContext := func(ctx context.Context, a int) {
 // get supervisor's id.
 supId := ctx.Value(easyworker.CTX_SUP_ID)
 // get child's id.
 childId := ctx.Value(easyworker.CTX_CHILD_ID)

 for i := 0; i < a; i++ {
  fmt.Println("Sup: ", supId, "Child:", childId, "counter:", i)
  time.Sleep(time.Millisecond)
 }
}

// create supervisor with context.
sup := NewSupervisorWithContext(context.Background())

// add child.
sup.NewChild(easyworker.NO_RESTART, loopWithContext, 10)
```

### EasyTask

This is simple way to run parallel tasks.
User doesn't need to manage goroutine, channel, ...
Number of workers is number of goroutine will run tasks.

In retry case, worker will re-use last parameters of task.

Result of each task is a []any.
You need to get true value from any(interface{}).

EasyTask example:

```go
// simple task.
func sum(a ...int) (ret int) {
 for _, i := range a {
  ret += i
 }
 return ret
}

func parallelTasks() {
 // number of workers.
 numWorkers := 3

 // retry times.
 retryTimes := 0

 // sleep time before re-run.
 retrySleep := 0

 // new config for EasyTask.
 config, _ := easyworker.NewConfig(sum, numWorkers, retryTimes, retrySleep)

 // new EasyTask.
 myTask, _ := easyworker.NewTask(config)

 // add tasks.
 myTask.AddTask(1, 2, 3)
 myTask.AddTask(3, 4, 5, 6, 7)
 myTask.AddTask(11, 22)

 // start workers and get results.
 r, e := myTask.Run()

 if e != nil {
  fmt.Println("run task failed, ", e)
 } else {
  fmt.Println("task result:", r)
 }
}
```

### EasyStream

This type is used for streaming type.
In this case, tasks are continuously send to worker by user's channel.
Results will receive from other channle of user.
Number of workers is number of goroutines used for processing streaming task.

In retry case, workers will re-use last parameters of task.

Result of each task is a []any.
You need to get true value from any(interface{}).

EasyStream example:

```go
// fun will do task
fnStr := func(a int, suffix string) string {
 if a%3 == 0 {
  panic("panic from user func")
 }
 return fmt.Sprintf("%d_%s", a, suffix)
}

num := easyworker.DefaultNumWorkers()

// input channel.
inCh := make(chan []any, num)

// result channel.
outCh := make(chan any, num)

// number of workers = number of cpu cores (logical cores).
config, _ := easyworker.NewConfig(fnStr, num, 3, 1000)

// test with stream.
myStream, _ := easyworker.NewStream(config, inCh, outCh)

// start stream.
myStream.Run()

// receive data from stream.
go func() {
  for {
    r := <-outCh
    fmt.Println("stream result: ", r)
  }
}()

// send data to stream.
go func() {
  for i := 0; i < 15; i++ {
    input := []any{i, "hello"}
    inCh <- input
    fmt.Println("stream sent: ", input)
  }
}()


// do something.

// stop all worker.
myStream.Stop()
```

### Monitor Go

A wrapper for goroutine for easy to monitor when goroutine was panic or run task done.

`Monitor` function will return two params.
First param is unique reference id.
Second param is channel that user can receive signal.

Signal is a struct with reference id and kind of end (failed, done).

If you need get result from last run, please call `GetResult`.

Example 1:

```go
loop := func(a int) {
 for i := 0; i < a; i++ {
  time.Sleep(time.Second)
  fmt.Println("loop at", i)
 }
 fmt.Println("Loop exit...")
}

// create go task.
g,_ := easyworker.NewGo(loop, 5)

go func() {
 // get a monitor to g.
 refId, ch := g.Monitor()

 // get a signal when g done/failed.
 sig := <-ch

 fmt.Println("ref:", refId, "ok")
}()

// start Go task.
g.Run()
```

Example 2:

```go
// create go task.
g,_ := easyworker.NewGo(loop, 5)

// get a monitor to g.
_, ch := g.Monitor()

// start Go task.
g.Run()

// task done
sig := <-ch

if sig.Signal != easyWorker.SIGNAL_DONE {
 // remove monitor link to Go.
 g.Demonitor()

 // retry one more.
 g.Run()
}
```

For other APIs please go to [pkg.go](https://pkg.go.dev/github.com/manhvu/easyworker)
