package main

import (
	"log"
	"time"
)

func Go(f func()) {
	go func() {
		// defer recover 捕获panic
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %+v", err)
			}
		}()

		f()
	}()
}

func main() {
	f := func() {
		panic("xxx")
	}
	Go(f)

	time.Sleep(1 * time.Second)
}
