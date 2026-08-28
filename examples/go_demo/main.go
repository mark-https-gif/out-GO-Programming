package main

import "fmt"

func add(a, b int) int {
	return a + b
}

func main() {
	fmt.Println("Hello from Go!")
	fmt.Println("3 + 4 =", add(3, 4))

	for i := 1; i < 5; i++ {
		fmt.Println("i =", i)
	}
}
