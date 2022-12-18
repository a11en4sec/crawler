package engine

import "github.com/a11en4sec/crawler/collect"

func NewEngine(opts ...Option) *Crawler {
	options := defaultOptions
	// 选项模式，根据需要丰富defaultOptions
	for _, opt := range opts {
		opt(&options)
	}
	e := &Crawler{}
	e.Visited = make(map[string]bool, 100)
	e.out = make(chan collect.ParseResult)
	e.failures = make(map[string]*collect.Request)
	e.options = options

	return e

}
