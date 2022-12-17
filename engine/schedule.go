package engine

import (
	"github.com/a11en4sec/crawler/collect"
	"go.uber.org/zap"
)

type Schedule struct {
	requestCh chan *collect.Request
	workerCh  chan *collect.Request
	out       chan collect.ParseResult
	options
}

func NewSchedule(opts ...Option) *Schedule {
	options := defaultOptions
	// 选项模式，根据需要丰富defaultOptions
	for _, opt := range opts {
		opt(&options)
	}
	s := &Schedule{}
	s.options = options
	return s
}

func (s *Schedule) Run() {
	requestCh := make(chan *collect.Request)
	workerCh := make(chan *collect.Request)
	out := make(chan collect.ParseResult)

	s.requestCh = requestCh
	s.workerCh = workerCh
	s.out = out

	go s.Schedule()

	for i := 0; i < s.WorkCount; i++ {
		go s.CreateWork()
	}

	s.HandleResult()
}

func (s *Schedule) Schedule() {
	var reqQueue []*collect.Request
	for _, seedTask := range s.Seeds {
		seedTask.RootReq.Task = seedTask
		seedTask.RootReq.Url = seedTask.Url

		reqQueue = append(reqQueue, seedTask.RootReq)
	}

	go func() {
		for {
			var req *collect.Request
			var ch chan *collect.Request
			//ch := make(chan *collect.Request)

			if len(reqQueue) > 0 {
				req = reqQueue[0]
				reqQueue = reqQueue[1:]
				ch = s.workerCh
			}

			select {
			case r := <-s.requestCh: // 监听一次fetch的解析结果中是否有新的request加入
				reqQueue = append(reqQueue, r)
			case ch <- req: // 传递给workerCh
			}
		}
	}()
}

func (s *Schedule) CreateWork() {
	for {
		r := <-s.workerCh
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

func (s *Schedule) HandleResult() {
	for {
		select {
		case result := <-s.out:
			for _, req := range result.Requesrts {
				s.requestCh <- req
			}

			for _, item := range result.Items {
				// todo: store
				s.Logger.Sugar().Info("get result ", item)
			}
		}
	}
}
