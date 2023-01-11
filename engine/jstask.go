package engine

import (
	"github.com/a11en4sec/crawler/spider"
	"github.com/robertkrimen/otto"
)

// AddJsReqs 用于动态规则添加请求。
func AddJsReqs(jreqs []map[string]interface{}) []*spider.Request {
	reqs := make([]*spider.Request, 0)

	for _, jreq := range jreqs {
		req := &spider.Request{}
		u, ok := jreq["URL"].(string)

		if !ok {
			return nil
		}

		req.URL = u
		req.RuleName, _ = jreq["RuleName"].(string)
		req.Method, _ = jreq["Method"].(string)
		req.Priority, _ = jreq["Priority"].(int64)
		reqs = append(reqs, req)
	}

	return reqs
}

// AddJsReq 用于动态规则添加请求。
func AddJsReq(jreq map[string]interface{}) []*spider.Request {
	reqs := make([]*spider.Request, 0)
	req := &spider.Request{}
	u, ok := jreq["URL"].(string)

	if !ok {
		return nil
	}

	req.URL = u
	req.RuleName, _ = jreq["RuleName"].(string)
	req.Method, _ = jreq["Method"].(string)
	req.Priority, _ = jreq["Priority"].(int64)
	reqs = append(reqs, req)

	return reqs
}

func (c *CrawlerStore) AddJSTask(m *spider.TaskModle) {
	task := &spider.Task{
		//Property: m.Property,
	}

	task.Rule.Root = func() ([]*spider.Request, error) {
		vm := otto.New()
		if err := vm.Set("AddJsReq", AddJsReqs); err != nil {
			return nil, err
		}

		v, err := vm.Eval(m.Root)

		if err != nil {
			return nil, err
		}

		e, err := v.Export()

		if err != nil {
			return nil, err
		}

		return e.([]*spider.Request), nil
	}

	for _, r := range m.Rules {
		// 将js编写的字符串，转成go代码，并执行
		paesrFunc := func(parse string) func(ctx *spider.Context) (spider.ParseResult, error) {
			return func(ctx *spider.Context) (spider.ParseResult, error) {
				vm := otto.New()

				if err := vm.Set("ctx", ctx); err != nil {
					return spider.ParseResult{}, err
				}

				v, err := vm.Eval(parse)

				if err != nil {
					return spider.ParseResult{}, err
				}

				e, err := v.Export()

				if err != nil {
					return spider.ParseResult{}, err
				}

				if e == nil {
					return spider.ParseResult{}, err
				}

				return e.(spider.ParseResult), err
			}
		}(r.ParseFunc)

		if task.Rule.Trunk == nil {
			task.Rule.Trunk = make(map[string]*spider.Rule, 0)
		}

		task.Rule.Trunk[r.Name] = &spider.Rule{
			ParseFunc: paesrFunc,
		}
	}

	c.Hash[task.Name] = task
	c.list = append(c.list, task)
}
