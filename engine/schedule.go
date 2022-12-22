package engine

import (
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
	requestCh   chan *collect.Request
	workerCh    chan *collect.Request
	priReqQueue []*collect.Request // 优先级队列
	reqQueue    []*collect.Request
	Logger      *zap.Logger
	//out       chan collect.ParseResult
	//options
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

//func (s *Schedule) Output() *collect.Request {
//	r := <-s.workerCh
//	return r
//}

func (s *Schedule) Schedule() {
	// 初始化不能放在协程中,否在会丢失请求
	var req *collect.Request
	var workerCh chan *collect.Request
	//ch := make(chan *collect.Request)

	//go func() {
	for {
		//var req *collect.Request
		//var ch chan *collect.Request
		//ch := make(chan *collect.Request)

		// 初始时，并且优先级队列不为空
		if req == nil && len(s.priReqQueue) > 0 {
			req = s.priReqQueue[0]
			s.priReqQueue = s.priReqQueue[1:]
			workerCh = s.workerCh
		}
		// 初始时，并且普通队列不为空
		if req == nil && len(s.reqQueue) > 0 {
			req = s.reqQueue[0]
			s.reqQueue = s.reqQueue[1:]
			workerCh = s.workerCh
		}

		select {
		case r := <-s.requestCh: // 种子req加入，以及fetch的解析结果中有新的request加入
			//s.reqQueue = append(s.reqQueue, r)
			if r.Priority > 0 {
				s.priReqQueue = append(s.priReqQueue, r)
			} else {
				s.reqQueue = append(s.reqQueue, r)
			}
		case workerCh <- req: // 传递给workerCh
			//fmt.Println("123")
			req = nil
			// todo: ？？ 为什么置nil
			workerCh = nil
		}
	}
	//}()

}
