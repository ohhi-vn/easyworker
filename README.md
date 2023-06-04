[![Go Reference](https://pkg.go.dev/badge/github.com/manhvu/easyworker.svg)](https://pkg.go.dev/github.com/manhvu/easyworker)

# Introduce

A Golang package for supporting worker supervisor model.
The package help developer easy to run tasks.
It's scalable with minimum effort.

easyworker inspired by Erlang OTP.

# Design

The package has two main part:

* The supervisor: Manage child, give a task to child & collect result.
* The child: Receive task, run task then send result to supervisor.

## Supervisor

Start worker and moniter worker.
Send task to worker and get result.
Restart child if it failed. It's depended about restart strategy of child.
Supervisor has one owner goroutine for send & manage signal to children.

## Child

Run task with user's function and handle error.
If user's function panic worker will check retry config and re-run if needed.
Each child has owner goroutine to run task.

# Guide

easywork support 2 type of worker and a type of supervisor:

* Task, Add a list of task and run worker.
* Stream, Start worker then push data to worker from channel.
* Supervisor, Start a supervisor for custom worker.

## EasyTask

This is simple way to run parallel task.
User doesn't need to manage goroutine, channel,...
Number of workers is number of goroutine will run tasks.

EasyTask example:

```go
fnSum = func(a ...int) int {
	sum := 0
	for _, i := range a {
		sum += i
	}
	return sum
}

// number of workers
numWorkers := 3

// retry times
retryTimes := 0

// sleep time before re-run
retrySleep := 0

// new config for EasyTask
config, _ := easyworker.NewConfig(fnSum, numWorkers, retryTimes, retrySleep)

// new EasyTask
task, _ := easyworker.NewTask(config)

// add tasks
myTask.AddTask(1, 2, 3)
myTask.AddTask(3, 4, 5, 6, 7)

// start workers
r, e := myTask.Run()

if e != nil {
    t.Error("run task failed, ", e)
} else {
    fmt.Println("task result:", r)
}
```

## EasyStream

This is used for streaming type.
In this case, tasks are continuously send to worker by user's channel.
Results will receive from other channle of user.
Number of workers is number of goroutines used for running stream task.

EasyStream example:

```go
// fun will do task
fnStr = func (a int, suffix string) string {
	if a%3 == 0 {
		panic("panic from user func")
	}
	return fmt.Sprintf("%d_%s", a, suffix)
}

inCh := make(chan []interface{})
outCh := make(chan interface{})

// number of workers = number of cpu cores (logical cores)
config, _ := easyworker.NewConfig(fnSum, easyworker.DefaultNumWorker(), 3, 1000)

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
        input := []interface{}{i, "3"}
        inCh <- input
        fmt.Println("stream sent: ", input)

    }
}()


...

// stop all worker
myStream.Stop()
```

## Supervisor

This is used for generic purpose worker(child).
Every children has a owner restart strategy.

Currently, child has three type of restart strategy:

* ALWAYS_RESTART, supervisor always restart child if it panic/done.
* NORMAL_RESTART, supervisor will restart if child was panic.
* NO_RESTART, supervisor will don't restart for any reason.

Child will be started after add to supervisor.

supervisor example:

```go
// example function need to run in child.
loopRun1 = func(a int) {
	for i := 0; i < a; i++ {
		time.Sleep(time.Second)
		fmt.Println("loop at", i)
	}
	fmt.Println("Loop exit..")
}

// example function run in child. It will panic if counter > 3
LoopRunWithPanic = func(a int) {
	for i := 0; i < a; i++ {
		time.Sleep(time.Second)
		fmt.Println("loop at", i)
		if i > 3 {
			panic("test loop with panic")
		}
	}
    // maybe you won't see this.
	fmt.Println("Loop exit..")
}

// create a supervisor
sup := easyworker.NewSupervisor()

// add direct child to supervisor.
sup.NewChild(easyworker.NORMAL_RESTART, LoopRun, 5)

// create a child
child, _ := easyworker.NewChild(easyworker.ALWAYS_RESTART, LoopRunWithPanic, 5)

// add exists child.
sup.AddChild(&child)

...

// stop all worker.
// this function depends how long fun return.
sup.Stop()
```
