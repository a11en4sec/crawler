package spider

import "regexp"

// 规则都是字符串了，这和之前的静态规则引擎有本质的不同,所以单独出来
type (
	TaskModle struct {
		Property
		Root  string      `json:"root_script"`
		Rules []RuleModle `json:"rule"`
	}
	RuleModle struct {
		Name      string `json:"name"`
		ParseFunc string `json:"parse_script"`
	}
)

func (c *Context) ParseJSReg(name string, reg string) ParseResult {
	re := regexp.MustCompile(reg)

	matches := re.FindAllSubmatch(c.Body, -1)
	result := ParseResult{}

	for _, m := range matches {
		u := string(m[1])

		result.Requesrts = append(
			result.Requesrts, &Request{
				Method:   "GET",
				Task:     c.Req.Task,
				URL:      u,
				Depth:    c.Req.Depth + 1,
				RuleName: name,
			})
	}

	return result
}

func (c *Context) OutputJS(reg string) ParseResult {
	re := regexp.MustCompile(reg)
	if ok := re.Match(c.Body); !ok {
		return ParseResult{
			Items: []interface{}{},
		}
	}

	result := ParseResult{
		Items: []interface{}{c.Req.URL},
	}

	return result
}
