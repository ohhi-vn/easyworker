package main

import (
	"fmt"
	"time"
)

func main() {
	fn := func(a int, b int) int {
		if a == 5 {
			panic("test panic")
		}
		return a + b
	}

	// inputs := make([][]int, 0)
	// inputs = append(inputs, []int{1, 2})
	// inputs = append(inputs, []int{3, 4})

	// StartSuper(inputs, fn, 1)

	inCh := make(chan []interface{})
	outCh := make(chan []interface{})

	fn(1, 2)

	StartSuperStream(inCh, outCh, fn, 1)

	go func() {
		for {
			r := <-outCh
			fmt.Println("result: ", r)
		}
	}()

	go func() {
		for i := 0; i < 10; i++ {
			input := []interface{}{i, 2}
			inCh <- input
			fmt.Println("sent: ", input)

		}
	}()

	time.Sleep(2 * time.Second)
}
