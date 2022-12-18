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
	failures    map[string]*collect.Request // 失败请求id -> 失败请求
	failureLock sync.Mutex
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
		r := s.scheduler.Pull() // 从workerChan中拉取一个req

		//判断r是否超过爬取的最高深度
		if err := r.Check(); err != nil {
			s.Logger.Error("check failed", zap.Error(err))
			continue
		}

		// 判断当前请求是否已被访问
		if !r.Task.Reload && s.HasVisited(r) {
			s.Logger.Debug("request has visited",
				zap.String("url:", r.Url),
			)
			continue
		}
		// 没有被访问，存到map中
		s.StoreVisited(r)

		body, err := s.Fetcher.Get(r)
		//fmt.Println("[++]", string(body))
		if err != nil {
			s.Logger.Error("can not fetch ",
				zap.Error(err),
				zap.String("url", r.Url),
			)
			// 请求失败，将请求放到错误map中
			s.SetFailure(r)
			continue
		}

		if len(body) < 6000 {
			s.Logger.Error("can't fetch ",
				zap.Int("length", len(body)),
				zap.String("url", r.Url),
			)
			// 请求失败，将请求放到错误map中
			s.SetFailure(r)
			continue
		}

		// todo: 错误的r，需要重新爬取
		result := r.ParseFunc(body, r)

		// 解析结果里面新的url，继续爬
		if len(result.Requesrts) > 0 {
			go s.scheduler.Push(result.Requesrts...)
		}

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

func (e *Crawler) HasVisited(r *collect.Request) bool {
	e.VisitedLock.Lock()
	defer e.VisitedLock.Unlock()
	unique := r.Unique()
	return e.Visited[unique]
}

func (e *Crawler) StoreVisited(reqs ...*collect.Request) {
	e.VisitedLock.Lock()
	defer e.VisitedLock.Unlock()

	for _, r := range reqs {
		unique := r.Unique()
		e.Visited[unique] = true
	}
}

func (e *Crawler) SetFailure(req *collect.Request) {
	// 没有被重新爬取过,第一次fetch失败
	if !req.Task.Reload {
		e.VisitedLock.Lock()
		unique := req.Unique()
		// 从已爬取的map中删除该req
		delete(e.Visited, unique)
		e.VisitedLock.Unlock()
	}
	e.failureLock.Lock()
	defer e.failureLock.Unlock()

	// 失败队列中没有，说明是首次失败
	if _, ok := e.failures[req.Unique()]; !ok {
		// 首次失败时，再重新执行一次
		e.failures[req.Unique()] = req
		e.scheduler.Push(req)
	}
	// todo: 失败2次，加载到失败队列中
}
