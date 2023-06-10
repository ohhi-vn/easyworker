package easyworker

import (
	"fmt"
	"testing"
)

func add(a int, b int) int {
	return a + b
}

func addWithPanic(a int, b int) int {
	if a%3 == 0 {
		panic("panic from user func")
	}
	return a + b
}

func sum(a ...int) int {
	sum := 0
	for _, i := range a {
		sum += i
	}
	return sum
}

func defaultConfig(fun any) Config {
	config, _ := NewConfig(fun, 1, 0, 0)
	return config
}

func strId(a int, suffix string) string {
	if a%3 == 0 {
		panic("panic from user func")
	}
	return fmt.Sprintf("%d_%s", a, suffix)
}

func TestIsNotFunc(t *testing.T) {
	_, err := NewConfig("fun", 1, 0, 0)

	if err == nil {
		t.Error("missed check function, ", err)
	}
}

func TestIncorrectNumWorker(t *testing.T) {
	_, err := NewConfig(add, 0, 0, 0)

	if err == nil {
		t.Error("incorrect number of worker is passed, ", err)
	}
}

func TestIncorrectNumRetry(t *testing.T) {
	_, err := NewConfig(add, 1, -1, 0)
	if err == nil {
		t.Error("incorrect number of retry is passed, ", err)
	}
}

func TestIncorrectRetryTime(t *testing.T) {
	_, err := NewConfig(add, 1, 1, -1)
	if err == nil {
		t.Error("incorrect retry time is passed, ", err)
	}
}
