package spider

import (
	"sync"
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

type TaskConfig struct {
	Name     string
	Cookie   string
	WaitTime int64
	Reload   bool
	MaxDepth int64
	Fetcher  string
	Limits   []LimitCofig
}

type LimitCofig struct {
	EventCount int
	EventDur   int // 秒
	Bucket     int // 桶大小
}

// Task 一个任务实例
type Task struct {
	//Property
	Visited     map[string]bool
	VisitedLock sync.Mutex
	//RootReq     *Request // 起始待爬的资源(seed)
	Rule RuleTree //规则树
	Options
}

func NewTask(opts ...Option) *Task {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}

	d := &Task{}
	d.Options = options

	return d
}

type Fetcher interface {
	Get(url *Request) ([]byte, error)
}
