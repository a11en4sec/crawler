package spider

import (
	"time"
)

type Context struct {
	Body []byte
	Req  *Request
}

func (c *Context) GetRule(ruleName string) *Rule {
	return c.Req.Task.Rule.Trunk[ruleName]
}

func (c *Context) Output(data interface{}) *DataCell {
	res := &DataCell{Task: c.Req.Task}
	res.Data = make(map[string]interface{})
	res.Data["Task"] = c.Req.Task.Name
	res.Data["Rule"] = c.Req.RuleName
	res.Data["Data"] = data
	res.Data["URL"] = c.Req.URL
	res.Data["Time"] = time.Now().Format("2006-01-02 15:04:05")

	return res
}
