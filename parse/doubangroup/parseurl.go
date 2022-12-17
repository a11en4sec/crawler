package doubangroup

import (
	"fmt"
	"github.com/a11en4sec/crawler/collect"
	"regexp"
	"strings"
)

func checkAntiCrawler(contents []byte) {
	// 检测是否被反扒拒绝--------
	//buf := new(bytes.Buffer)
	//buf.ReadFrom(closer)
	//respStr := buf.String()
	if strings.Contains(string(contents), "有异常请求从你的") {
		fmt.Println("[---]:", "触发反爬")
	}
	// ----------------------

}

const urlListRe = `(https://www.douban.com/group/topic/[0-9a-z]+/)"[^>]*>([^<]+)</a>`

func ParseURL(contents []byte, req *collect.Request) collect.ParseResult {

	// 是否触发反爬
	checkAntiCrawler(contents)

	//fmt.Println("[body]:", string(contents))
	re := regexp.MustCompile(urlListRe)

	matches := re.FindAllSubmatch(contents, -1)
	result := collect.ParseResult{}

	//var fetchUrlCount int

	for _, m := range matches {
		u := string(m[1])
		//fmt.Printf("[+ fetchUrlCount:%d] %s\n]", fetchUrlCount, u)
		result.Requesrts = append(
			result.Requesrts, &collect.Request{
				Url:      u,
				WaitTime: req.WaitTime,
				Cookie:   req.Cookie,
				Depth:    req.Depth + 1,
				ParseFunc: func(c []byte, request *collect.Request) collect.ParseResult {
					return GetContent(c, u)
				},
			})
	}
	return result
}

const ContentRe = `<div class="topic-content">[\s\S]*?阳台[\s\S]*?<div`

func GetContent(contents []byte, url string) collect.ParseResult {
	re := regexp.MustCompile(ContentRe)

	ok := re.Match(contents)
	if !ok {
		return collect.ParseResult{
			Items: []interface{}{},
		}
	}

	result := collect.ParseResult{
		Items: []interface{}{url},
	}

	return result
}
