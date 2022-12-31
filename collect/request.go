package collect

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/a11en4sec/crawler/limiter"
	"github.com/a11en4sec/crawler/storage"
)

type Property struct {
	Name   string `json:"name"` // 任务名称，应保证唯一性
	URL    string `json:"url"`
	Cookie string `json:"cookie"`
	//WaitTime time.Duration `json:"wait_time"`
	WaitTime int64 `json:"wait_time"`
	Reload   bool  `json:"reload"` // 网站是否可以重复爬取
	MaxDepth int64 `json:"max_depth"`
}

// Task 一个任务实例
type Task struct {
	Property
	Visited     map[string]bool
	VisitedLock sync.Mutex
	//RootReq     *Request // 起始待爬的资源(seed)
	Rule    RuleTree //规则树
	Fetcher Fetcher
	Storage storage.Storage     // 存储
	Limit   limiter.RateLimiter // 限速器
}

// Request 单个请求
type Request struct {
	Task     *Task
	URL      string
	Method   string
	Priority int64
	Depth    int64
	//ParseFunc func([]byte, *Request) ParseResult
	RuleName string
	TmpData  *Temp // 方法数据

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

// Unique 请求的唯一识别码
func (r *Request) Unique() string {
	block := md5.Sum([]byte(r.URL + r.Method))

	return hex.EncodeToString(block[:])
}

func (r *Request) Fetch() ([]byte, error) {
	if err := r.Task.Limit.Wait(context.Background()); err != nil {
		return nil, err
	}

	// 随机休眠，模拟人类行为
	st := rand.Int63n(r.Task.WaitTime * 1000)
	time.Sleep(time.Duration(st) * time.Millisecond)

	return r.Task.Fetcher.Get(r)
}
