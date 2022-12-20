package collect

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

type Property struct {
	Name     string        `json:"name"` // 任务名称，应保证唯一性
	Url      string        `json:"url"`
	Cookie   string        `json:"cookie"`
	WaitTime time.Duration `json:"wait_time"`
	Reload   bool          `json:"reload"` // 网站是否可以重复爬取
	MaxDepth int64         `json:"max_depth"`
}

// Task 一个任务实例
type Task struct {
	//Name        string // 用户界面显示的名称（应保证唯一性）
	//Url         string
	//Cookie      string
	//WaitTime    time.Duration
	//Reload      bool // 网站是否可以重复爬取
	//MaxDepth    int
	Property
	Visited     map[string]bool
	VisitedLock sync.Mutex
	//RootReq     *Request // 起始待爬的资源(seed)
	Rule    RuleTree //规则树
	Fetcher Fetcher
}

type Context struct {
	Body []byte
	Req  *Request
}

// Request 单个请求
type Request struct {
	Task     *Task
	Url      string
	Method   string
	Priority int64
	Depth    int64
	//ParseFunc func([]byte, *Request) ParseResult
	RuleName string
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
	block := md5.Sum([]byte(r.Url + r.Method))
	return hex.EncodeToString(block[:])
}
