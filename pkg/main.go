package main

func main() {
	fn := func(a int, b int) int {
		return a + b
	}

	StartSuper(fn, 3)
}
