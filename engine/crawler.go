package engine

import (
	"github.com/a11en4sec/crawler/collect"
	"github.com/a11en4sec/crawler/parse/doubangroup"
	"go.uber.org/zap"
	"sync"
)

func init() {
	Store.Add(doubangroup.DoubangroupTask)
}

func (c *CrawlerStore) Add(task *collect.Task) {
	c.hash[task.Name] = task
	c.list = append(c.list, task)
}

// Store 全局蜘蛛种类实例
var Store = &CrawlerStore{
	list: []*collect.Task{},
	hash: map[string]*collect.Task{},
}

type CrawlerStore struct {
	list []*collect.Task
	hash map[string]*collect.Task
}

type Crawler struct {
	out         chan collect.ParseResult
	Visited     map[string]bool // keys is md5(URL + method)
	VisitedLock sync.Mutex
	failures    map[string]*collect.Request // 失败请求id -> 失败请求
	failureLock sync.Mutex
	options
}

func (s *Crawler) Run() {
	go s.Schedule()

	for i := 0; i < s.WorkCount; i++ {
		go s.CreateWork()
	}

	s.HandleResult()
}

func (s *Crawler) Schedule() {
	var reqQueue []*collect.Request
	for _, seedTask := range s.Seeds {
		//seedTask.RootReq.Task = seedTask
		//seedTask.RootReq.Url = seedTask.Url
		//reqQueue = append(reqQueue, seedTask.RootReq)

		// 从全局store中取出Task
		task := Store.hash[seedTask.Name]
		task.Fetcher = seedTask.Fetcher

		// 取出Task的根，根节点(执行入口),生成爬虫的种子网站
		rootReqs := task.Rule.Root()
		for _, req := range rootReqs {
			req.Task = task
		}
		reqQueue = append(reqQueue, rootReqs...)
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
		req := s.scheduler.Pull() // 从workerChan中拉取一个req

		//判断r是否超过爬取的最高深度
		if err := req.Check(); err != nil {
			s.Logger.Error("check failed", zap.Error(err))
			continue
		}

		// 判断当前请求是否已被访问
		if !req.Task.Reload && s.HasVisited(req) {
			s.Logger.Debug("request has visited",
				zap.String("url:", req.Url),
			)
			continue
		}
		// 没有被访问，存到map中
		s.StoreVisited(req)

		body, err := s.Fetcher.Get(req)
		//fmt.Println("[++]", string(body))
		if err != nil {
			s.Logger.Error("can not fetch ",
				zap.Error(err),
				zap.String("url", req.Url),
			)
			// 请求失败，将请求放到错误map中
			s.SetFailure(req)
			continue
		}

		if len(body) < 6000 {
			s.Logger.Error("can't fetch ",
				zap.Int("length", len(body)),
				zap.String("url", req.Url),
			)
			// 请求失败，将请求放到错误map中
			s.SetFailure(req)
			continue
		}

		//result := req.ParseFunc(body, req)
		// //获取当前任务对应的规则 去处理fetch(req)回来的结果
		rule := req.Task.Rule.Trunk[req.RuleName]

		// 处理fetch(req)回来的结果
		result := rule.ParseFunc(&collect.Context{
			body,
			req,
		})
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

func (s *Crawler) HasVisited(r *collect.Request) bool {
	s.VisitedLock.Lock()
	defer s.VisitedLock.Unlock()
	unique := r.Unique()
	return s.Visited[unique]
}

func (s *Crawler) StoreVisited(reqs ...*collect.Request) {
	s.VisitedLock.Lock()
	defer s.VisitedLock.Unlock()

	for _, r := range reqs {
		unique := r.Unique()
		s.Visited[unique] = true
	}
}

func (s *Crawler) SetFailure(req *collect.Request) {
	// 没有被重新爬取过,第一次fetch失败
	if !req.Task.Reload {
		s.VisitedLock.Lock()
		unique := req.Unique()
		// 从已爬取的map中删除该req
		delete(s.Visited, unique)
		s.VisitedLock.Unlock()
	}
	s.failureLock.Lock()
	defer s.failureLock.Unlock()

	// 失败队列中没有，说明是首次失败
	if _, ok := s.failures[req.Unique()]; !ok {
		// 首次失败时，再重新执行一次
		s.failures[req.Unique()] = req
		s.scheduler.Push(req)
	}
	// todo: 失败2次，加载到失败队列中
}
