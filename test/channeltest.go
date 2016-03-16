package main

import (
	"fmt"
	"time"
)

func Count(ch chan int, i int) {
	ch <- i
	fmt.Println("Counting")
}
func main() {
	chs := make([]chan int, 10)

	for i := 0; i < 10; i++ {
		chs[i] = make(chan int)
		go Count(chs[i], i)
	}

	for _, ch := range chs {
		i := <-ch
		fmt.Println("Receive value:", i)
	}

	//must wait, will echo 10 string
	time.Sleep(time.Second)
}
