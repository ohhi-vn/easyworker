# Introduce

A Golang package for supporting worker supervisor model.
The package help developer easy to run tasks.
It's scalable with minimum effort.

# Design

The package has two main part:

* The supervisor: Manage worker, give a task to worker & collect result.
* The worker: Receive task, run task then send result to supervisor.

```mermain

```

# Guide

easywork support 2 type of worker:

* Task. Add a list of task and run worker
* Stream. Start worker then push data to worker from channel

## EasyTask



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

task, err := NewTask(fnSum, 3, 1)

if err != nil {
    t.Error("cannot create task, ", err)
    return
}

task.AddTask(1, 2, 3)
task.AddTask(3, 4, 5, 6, 7)

r, e := eWorker.Run()
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

time.Sleep(time.Millisecond)

// test with stream.
eWorker, err := NewStream(inCh, outCh, fnStr, 2, 1)
if err != nil {
    t.Error("create EasyWorker failed, ", err)
}

e := eWorker.Run()
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

fmt.Println("send stop signal to stream")
eWorker.StopStream()
```