package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(id int) {
	fmt.Printf("worker %d starting \n ", id)
	time.Sleep(time.Second * 1)
	fmt.Printf("worker %d done \n ", id)

}

func main() {

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		i := i

		go func() {
			defer wg.Done()
			worker(i)
		}()
	}

	wg.Wait()

}
