package main

import "fmt"

// 第一个阶段，数字的生成器
func Generate(ch chan<- int) {
	for i := 2; ; i++ {
		ch <- i // Send 'i' to channel 'ch'.
	}
}

// 筛选，排除不能够被prime整除的数
func Filter(in <-chan int, out chan<- int, prime int) {
	for {
		i := <-in // 获取上一个阶段的
		if i%prime != 0 {
			out <- i // Send 'i' to 'out'.
		}
	}
}

func main() {
	ch := make(chan int)
	go Generate(ch)
	for i := 0; i < 100; i++ {
		prime := <-ch // 获取上一个阶段输出的第一个数，其必然为素数
		fmt.Println(prime)
		ch1 := make(chan int)
		go Filter(ch, ch1, prime)
		ch = ch1 // 前一个阶段的输出作为后一个阶段的输入。
	}
}
