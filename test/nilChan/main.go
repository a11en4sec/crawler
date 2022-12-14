package main

import "fmt"

func main() {
	//var ch chan *int
	ch := make(chan *int)
	go func() {
		<-ch
	}()
	select {
	case ch <- nil:
		fmt.Println("it's time")
	}
}
