package main

import (
	"fmt"
	"sync/atomic"
	"time"
)

var flag int64 = 0
var count int64 = 0

func add() {
	for {
		if atomic.CompareAndSwapInt64(&flag, 0, 1) {
			count++
			atomic.StoreInt64(&flag, 0)
			fmt.Println("[+]count:", count)
			return
		}
	}
}
func main() {
	go add()
	go add()
	time.Sleep(time.Second * 2)
}
