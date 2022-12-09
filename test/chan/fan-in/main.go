package main

import (
	"fmt"
	"time"
)

func search(ch chan string, msg string) {
	var i int
	for {
		// 模拟找到了关键字
		ch <- fmt.Sprintf("get %s %d", msg, i)
		i++
		time.Sleep(1000 * time.Millisecond)
	}
}

func searchToNewChan(msg string) chan string {
	var ch = make(chan string)
	go func() {
		var i int
		for {
			ch <- fmt.Sprintf("get %s %d", msg, i)
			i++
			time.Sleep(time.Second * 1)
		}
	}()
	return ch
}

func search1() {
	ch := make(chan string)
	go search(ch, "jonson")
	go search(ch, "olaya")

	for i := range ch {
		fmt.Println(i)
	}
}

func main() {
	ch1 := searchToNewChan("a")
	ch2 := searchToNewChan("b")

	for {
		select {
		case msg := <-ch1:
			fmt.Println(msg)
		case msg := <-ch2:
			fmt.Println(msg)
		}
	}
}
