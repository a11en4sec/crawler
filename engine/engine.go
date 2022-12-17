package engine

import "github.com/a11en4sec/crawler/collect"

func NewEngine(opts ...Option) *Crawler {
	options := defaultOptions
	// 选项模式，根据需要丰富defaultOptions
	for _, opt := range opts {
		opt(&options)
	}
	e := &Crawler{}
	out := make(chan collect.ParseResult)
	e.out = out
	e.options = options

	return e

}
