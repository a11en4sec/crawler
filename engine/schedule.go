package engine

import (
	"github.com/a11en4sec/crawler/collect"
	"go.uber.org/zap"
)

type ScheduleEngine struct {
	requestCh chan *collect.Request
	workerCh  chan *collect.Request
	out       chan collect.ParseResult
	options
}

func (s *ScheduleEngine) Run() {
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

func (s *ScheduleEngine) Schedule() {
	var reqQueue = s.Seeds
	go func() {
		for {
			var req *collect.Request
			//var ch chan *collect.Request
			ch := make(chan *collect.Request)

			if len(reqQueue) > 0 {
				req = reqQueue[0]
				reqQueue = reqQueue[1:]
				ch = s.workerCh
			}

			select {
			case r := <-s.requestCh:
				reqQueue = append(reqQueue, r)
			case ch <- req:
			}
		}
	}()
}

func NewSchedule(opts ...Option) *ScheduleEngine {
	options := defaultOptions

	// 选项模式，根据需要丰富defaultOptions
	for _, opt := range opts {
		opt(&options)
	}
	s := &ScheduleEngine{}

	s.options = options

	return s
}

func (s *ScheduleEngine) CreateWork() {
	for {
		r := <-s.workerCh
		// 判断r是否超过爬取的最高深度
		if err := r.Check(); err != nil {
			s.Logger.Error("check failed", zap.Error(err))
			continue
		}
		body, err := s.Fetcher.Get(r)
		//fmt.Println("[++]", string(body))
		if err != nil {
			s.Logger.Error("can not fetch ", zap.Error(err))
			continue
		}
		// todo: 错误的r，需要重新爬取
		result := r.ParseFunc(body, r)
		s.out <- result
	}
}

func (s *ScheduleEngine) HandleResult() {
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
