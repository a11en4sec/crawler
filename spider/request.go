package spider

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"math/rand"
	"time"
)

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

func (r *Request) Fetch() ([]byte, error) {
	if err := r.Task.Limit.Wait(context.Background()); err != nil {
		return nil, err
	}

	// 随机休眠，模拟人类行为
	st := rand.Int63n(r.Task.WaitTime * 1000)
	time.Sleep(time.Duration(st) * time.Millisecond)

	return r.Task.Fetcher.Get(r)
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

	if r.Task.Closed {
		return errors.New("task has Closed")
	}

	return nil
}

// Unique 请求的唯一识别码
func (r *Request) Unique() string {
	block := md5.Sum([]byte(r.URL + r.Method))

	return hex.EncodeToString(block[:])
}
