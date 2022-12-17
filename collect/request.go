package collect

import (
	"errors"
	"time"
)

type Request struct {
	Url       string
	Cookie    string
	ParseFunc func([]byte, *Request) ParseResult
	WaitTime  time.Duration
	Depth     int
	MaxDepth  int
}

type ParseResult struct {
	Requesrts []*Request    // 用于进一步获取数据
	Items     []interface{} // 收到的的数据
}

func (r *Request) Check() error {
	if r.Depth > r.MaxDepth {
		return errors.New("max depth limit reached")
	}
	return nil
}
