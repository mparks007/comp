package main

import "fmt"

import "sync"

type x struct {
	a int
	b int
}

func main() {
	fmt.Println("Adder Client")

	wg := &sync.WaitGroup{}
	nums := make([]int, 10)

	for i := range nums {
		wg.Add(1)
		go func() {
			fmt.Println(addOne(i))
			wg.Done()
		}()
	}
	wg.Wait()
}

func addOne(i int) int {
	return i + 1
}
