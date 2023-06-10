package easyworker

import (
	"errors"
	"log"
	"testing"
)

func simpleLoopReturn(a int) int {
	ret := 0
	for i := 0; i < a; i++ {
		ret += i
	}

	return ret
}

func returnMultiValue() (i int, f float64, b bool, e error) {
	i = 123
	f = 1.2
	b = true
	e = errors.New("test return")
	return
}

func TestInvokeNoArg(t *testing.T) {
	result, err := invokeFun(simpleLoopNoArg)

	if err != nil {
		t.Error("test invoke with no argument failed, ", err)
	} else {
		log.Println("result: ", result)
	}
}

func TestInvokeIncorrectNumArg(t *testing.T) {
	_, err := invokeFun(simpleLoopWithPanic, 3, 3)

	if err == nil {
		t.Error("test invoke with incorrect argument failed")
	}
}

func TestInvokeIncorrectNumArg2(t *testing.T) {
	_, err := invokeFun(simpleLoopWithPanic)

	if err == nil {
		t.Error("test invoke with incorrect argument failed")
	}
}

func TestInvokeIncorrectNumArg3(t *testing.T) {
	_, err := invokeFun(simpleLoopWithPanic, "a")

	if err == nil {
		t.Error("test invoke with incorrect argument failed")
	}
}

func TestInvokePanic(t *testing.T) {
	_, err := invokeFun(simpleLoopWithPanic, 5)

	if err == nil {
		t.Error("expected error but no return error.")
	} else {
		log.Println("expected is ok, err: ", err)
	}
}

func TestInvokeReturn(t *testing.T) {
	result, err := invokeFun(simpleLoopReturn, 5)

	if err != nil {
		t.Error("test invoke with no argument failed, ", err)
	} else {
		log.Println("result: ", result)
		if result[0] != 10 {
			t.Error("return incorrect value")
		}
	}
}

func TestInvokeReturnMultiValue(t *testing.T) {
	result, err := invokeFun(returnMultiValue)

	if err != nil {
		t.Error("test invoke with no argument failed, ", err)
	} else {
		log.Println("result: ", result)
	}
}
