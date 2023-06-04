package easyworker

import (
	"fmt"
	"testing"
)

func Add(a int, b int) int {
	return a + b
}

func AddWithPanic(a int, b int) int {
	if a%3 == 0 {
		panic("panic from user func")
	}
	return a + b
}

func Sum(a ...int) int {
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

func StrId(a int, suffix string) string {
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
	_, err := NewConfig(Add, 0, 0, 0)

	if err == nil {
		t.Error("incorrect number of worker is passed, ", err)
	}
}

func TestIncorrectNumRetry(t *testing.T) {
	_, err := NewConfig(Add, 0, -1, 0)
	if err == nil {
		t.Error("incorrect number of retry is passed, ", err)
	}
}
