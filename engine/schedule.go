package engine

import (
	"fmt"
	"github.com/a11en4sec/crawler/collect"
	"go.uber.org/zap"
)

type Scheduler interface {
	Schedule()                // 方法负责启动调度器
	Push(...*collect.Request) // 将请求放入到调度器中
	Pull() *collect.Request   // 从调度器中获取请求
}

// Schedule 实现了接口Scheduler
type Schedule struct {
	requestCh chan *collect.Request
	workerCh  chan *collect.Request
	//out       chan collect.ParseResult
	//options
	reqQueue []*collect.Request
	Logger   *zap.Logger
}

func NewSchedule() *Schedule {
	s := &Schedule{}

	requestCh := make(chan *collect.Request)
	workerCh := make(chan *collect.Request)

	s.requestCh = requestCh
	s.workerCh = workerCh

	return s
}

func (s *Schedule) Push(reqs ...*collect.Request) {
	for _, req := range reqs {
		s.requestCh <- req
	}
}

func (s *Schedule) Pull() *collect.Request {
	r := <-s.workerCh
	return r
}

func (s *Schedule) Output() *collect.Request {
	r := <-s.workerCh
	return r
}

func (s *Schedule) Schedule() {
	go func() {
		for {
			var req *collect.Request
			var ch chan *collect.Request
			//ch := make(chan *collect.Request)

			if len(s.reqQueue) > 0 {
				req = s.reqQueue[0]
				s.reqQueue = s.reqQueue[1:]
				ch = s.workerCh
			}

			select {
			case r := <-s.requestCh: // 监听一次fetch的解析结果中是否有新的request加入
				s.reqQueue = append(s.reqQueue, r)
			case ch <- req: // 传递给workerCh
				fmt.Println("123")
			}
		}
	}()

}
