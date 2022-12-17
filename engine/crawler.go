package engine

import (
	"github.com/a11en4sec/crawler/collect"
	"go.uber.org/zap"
	"sync"
)

type Crawler struct {
	out         chan collect.ParseResult
	Visited     map[string]bool // keys is md5(URL + method)
	VisitedLock sync.Mutex
	options
}

func (e *Crawler) Run() {
	go e.Schedule()

	for i := 0; i < e.WorkCount; i++ {
		go e.CreateWork()
	}

	e.HandleResult()
}

func (s *Crawler) Schedule() {
	var reqQueue []*collect.Request
	for _, seedTask := range s.Seeds {
		seedTask.RootReq.Task = seedTask
		seedTask.RootReq.Url = seedTask.Url
		reqQueue = append(reqQueue, seedTask.RootReq)
	}
	go s.scheduler.Schedule()
	go s.scheduler.Push(reqQueue...)

	//go func() {
	//	for {
	//		var req *collect.Request
	//		var ch chan *collect.Request
	//		//ch := make(chan *collect.Request)
	//
	//		if len(reqQueue) > 0 {
	//			req = reqQueue[0]
	//			reqQueue = reqQueue[1:]
	//			ch = s.workerCh
	//		}
	//
	//		select {
	//		case r := <-s.requestCh: // 监听一次fetch的解析结果中是否有新的request加入
	//			reqQueue = append(reqQueue, r)
	//		case ch <- req: // 传递给workerCh
	//		}
	//	}
	//}()
}

func (s *Crawler) CreateWork() {
	for {
		r := s.scheduler.Pull()
		//判断r是否超过爬取的最高深度
		if err := r.Check(); err != nil {
			s.Logger.Error("check failed", zap.Error(err))
			continue
		}
		body, err := s.Fetcher.Get(r)
		//fmt.Println("[++]", string(body))
		if len(body) < 6000 {
			s.Logger.Error("can't fetch ",
				zap.Int("length", len(body)),
				zap.String("url", r.Url),
			)
			continue
		}

		if err != nil {
			s.Logger.Error("can not fetch ",
				zap.Error(err),
				zap.String("url", r.Url),
			)
			continue
		}
		// todo: 错误的r，需要重新爬取
		result := r.ParseFunc(body, r)
		s.out <- result
	}
}

func (s *Crawler) HandleResult() {
	for {
		select {
		case result := <-s.out:
			//for _, req := range result.Requesrts {
			//	s.requestCh <- req
			//}
			for _, item := range result.Items {
				// todo: store
				s.Logger.Sugar().Info("get result ", item)
			}
		}
	}
}
