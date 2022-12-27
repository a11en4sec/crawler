package main

import (
	"context"
	"fmt"
	"golang.org/x/time/rate"
	"time"
)

func main() {
	// 每隔 1 秒钟向桶中放入 1 个令牌
	//limit := rate.NewLimiter(rate.Limit(1), 2)

	// 每500毫秒放一个令牌
	limit := rate.NewLimiter(rate.Every(time.Millisecond*500), 2)
	for {
		if err := limit.Wait(context.Background()); err == nil {
			fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
		}
	}
}
