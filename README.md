# Introduce

A Golang package for supporting worker supervisor model.
The package help developer easy to run tasks.
It's scalable with minimum effort.

# Design

The package has two main part:

* The supervisor: Manage worker, give a task to worker & collect result.
* The worker: Receive task, run task then send result to supervisor.

## Supervisor

Start worker and moniter worker.
Send task to worker and get result.

## Worker

Run task with user's function and handle error.
If user's function panic worker will check retry config and re-run if needed.

# Guide

easywork support 2 type of worker:

* Task. Add a list of task and run worker
* Stream. Start worker then push data to worker from channel

## EasyTask

This is simple way to run parallel task.
User doesn't to manage goroutine, channel,...

## EasyStream

This is used for streaming type.
In this case, tasks are continuously send to worker by user's channel.
Results will receive from other channle of user.

# Example

EasyTask example:

```go
fnSum = func(a ...int) int {
	sum := 0
	for _, i := range a {
		sum += i
	}
	return sum
}

numWorkers := 3
retryTimes := 0
retrySleep := 0

config, _ := NewConfig(fnSum, numWorkers, retryTimes, retrySleep)

task, err := NewTask(config)

if err != nil {
    t.Error("cannot create task, ", err)
    return
}

myTask.AddTask(1, 2, 3)
myTask.AddTask(3, 4, 5, 6, 7)

r, e := myTask.Run()
if e != nil {
    t.Error("run task failed, ", e)
} else {
    fmt.Println("task result:", r)
}
```

EasyStream example:

```go
fnStr = func (a int, suffix string) string {
	if a%3 == 0 {
		panic("panic from user func")
	}
	return fmt.Sprintf("%d_%s", a, suffix)
}

inCh := make(chan []interface{}, 1)
outCh := make(chan interface{})

// number of workers = number of cpu cores (logical cores)
config, _ := NewConfig(fnSum, DefaultNumWorker(), 3, 1000)

// test with stream.
myStream, err := NewStream(config, inCh, outCh)
if err != nil {
    t.Error("create EasyWorker failed, ", err)
}

e := myStream.Run()
if e != nil {
    t.Error("run stream task failed, ", e)
} else {
    fmt.Println("stream is running")
}

go func() {
    for {
        r := <-outCh
        fmt.Println("stream result: ", r)
    }
}()

go func() {
    for i := 0; i < 15; i++ {
        input := []interface{}{i, "3"}
        inCh <- input
        fmt.Println("stream sent: ", input)

    }
}()

time.Sleep(2 * time.Second)

// Stop all worker
myStream.Stop()
```