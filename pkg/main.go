package main

import (
	"fmt"
	"time"
)

func main() {
	fn := func(a int, b int) int {
		if a%3 == 0 {
			panic("test panic")
		}
		return a + b
	}
	fn2 := func(a int, b string) string {
		if a%3 == 10 {
			panic("test panic")
		}
		return fmt.Sprintf("%d <> %s", a, b)
	}

	inputs := make([][]int, 0)
	inputs = append(inputs, []int{1, 2})
	inputs = append(inputs, []int{3, 4})

	StartSuper(inputs, fn, 1, 1)

	inCh := make(chan []interface{})
	outCh := make(chan []interface{})

	// test with stream.
	StartSuperStream(inCh, outCh, fn2, 2, 3)

	go func() {
		for {
			r := <-outCh
			fmt.Println("result: ", r)
		}
	}()

	go func() {
		for i := 0; i < 15; i++ {
			input := []interface{}{i, "3"}
			inCh <- input
			fmt.Println("sent: ", input)

		}
	}()

	time.Sleep(2 * time.Second)
}
