package collect

import (
	"errors"
	"sync"
	"time"
)

// Task 一个任务实例
type Task struct {
	Url         string
	Cookie      string
	WaitTime    time.Duration
	MaxDepth    int
	Visited     map[string]bool
	VisitedLock sync.Mutex
	RootReq     *Request // 起始待爬的资源(seed)
	Fetcher     Fetcher
}

// Request 单个请求
type Request struct {
	Task      *Task
	Url       string
	Depth     int
	ParseFunc func([]byte, *Request) ParseResult
}

type ParseResult struct {
	Requesrts []*Request    // 用于进一步获取数据
	Items     []interface{} // 收到的的数据
}

func (r *Request) Check() error {
	//fmt.Printf("r.depth:%d , r.Task.MaxDepth:%d\n", r.Depth, r.Task.MaxDepth)
	if r.Depth > r.Task.MaxDepth {
		return errors.New("max depth limit reached")
	}
	return nil
}
