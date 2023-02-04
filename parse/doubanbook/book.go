package doubanbook

import (
	"regexp"
	"strconv"

	"github.com/a11en4sec/crawler/spider"

	"go.uber.org/zap"
)

// DoubanBookTask 引擎初始化的时候加载
var DoubanBookTask = &spider.Task{
	Options: spider.Options{Name: "douban_book_list"},
	Rule: spider.RuleTree{
		Root: func() ([]*spider.Request, error) {
			roots := []*spider.Request{
				{
					Priority: 1,
					URL:      "https://book.douban.com",
					Method:   "GET",
					RuleName: "数据tag",
				},
			}

			return roots, nil
		},
		Trunk: map[string]*spider.Rule{
			"数据tag": {ParseFunc: ParseTag},
			"书籍列表":  {ParseFunc: ParseBookList},
			"书籍简介": {
				ItemFields: []string{
					"书名",
					"作者",
					"页数",
					"出版社",
					"得分",
					"价格",
					"简介",
				},
				ParseFunc: ParseBookDetail,
			},
		},
	},
}

const regexpStr = `<a href="([^"]+)" class="tag">([^<]+)</a>`

func ParseTag(ctx *spider.Context) (spider.ParseResult, error) {
	re := regexp.MustCompile(regexpStr)

	matches := re.FindAllSubmatch(ctx.Body, -1)
	result := spider.ParseResult{}

	if len(matches) == 0 {
		zap.L().Info(" regex ", zap.String("regex not match", "tag not match"))
	}

	for _, m := range matches {
		result.Requesrts = append(
			result.Requesrts, &spider.Request{
				Method:   "GET",
				Task:     ctx.Req.Task,
				URL:      "https://book.douban.com" + string(m[1]),
				Depth:    ctx.Req.Depth + 1,
				RuleName: "书籍列表",
			})
	}

	zap.S().Debugln("parse book tag,count:", len(result.Requesrts))

	// 在添加limit之前，临时减少抓取数量,防止被服务器封禁
	//result.Requesrts = result.Requesrts[:3]
	return result, nil
}

const BooklistRe = `<a.*?href="([^"]+)" title="([^"]+)"`

func ParseBookList(ctx *spider.Context) (spider.ParseResult, error) {
	re := regexp.MustCompile(BooklistRe)
	matches := re.FindAllSubmatch(ctx.Body, -1)
	result := spider.ParseResult{}

	for _, m := range matches {
		req := &spider.Request{
			Method:   "GET",
			Task:     ctx.Req.Task,
			URL:      string(m[1]),
			Depth:    ctx.Req.Depth + 1,
			RuleName: "书籍简介",
		}
		// 书名缓存到了临时的 tmp 结构中供下一个阶段读取
		req.TmpData = &spider.Temp{}
		if err := req.TmpData.Set("book_name", string(m[2])); err != nil {
			zap.L().Error("Set TmpData failed", zap.Error(err))
		}

		result.Requesrts = append(result.Requesrts, req)
		zap.S().Debugln("parse book list,count:", len(result.Requesrts), "url:", ctx.Req.URL)
	}
	// todo: 在添加limit之前，临时减少抓取数量,防止被服务器封禁
	//result.Requesrts = result.Requesrts[:3]

	return result, nil
}

var autoRe = regexp.MustCompile(`<span class="pl"> 作者</span>:[\d\D]*?<a.*?>([^<]+)</a>`)

var public = regexp.MustCompile(`<span class="pl">出版社:</span>[\d\D]*?<a.*?>([^<]+)</a>`)
var pageRe = regexp.MustCompile(`<span class="pl">页数:</span> ([^<]+)<br/>`)
var priceRe = regexp.MustCompile(`<span class="pl">定价:</span>([^<]+)<br/>`)
var scoreRe = regexp.MustCompile(`<strong class="ll rating_num " property="v:average">([^<]+)</strong>`)
var intoRe = regexp.MustCompile(`<div class="intro">[\d\D]*?<p>([^<]+)</p></div>`)

func ParseBookDetail(ctx *spider.Context) (spider.ParseResult, error) {
	bookName := ctx.Req.TmpData.Get("book_name")
	page, _ := strconv.Atoi(ExtraString(ctx.Body, pageRe))

	book := map[string]interface{}{
		"书名":  bookName,
		"作者":  ExtraString(ctx.Body, autoRe),
		"页数":  page,
		"出版社": ExtraString(ctx.Body, public),
		"得分":  ExtraString(ctx.Body, scoreRe),
		"价格":  ExtraString(ctx.Body, priceRe),
		"简介":  ExtraString(ctx.Body, intoRe),
	}
	// 将context 转成dataCell
	data := ctx.Output(book)

	result := spider.ParseResult{
		Items: []interface{}{data},
	}

	zap.S().Debugln("parse book detail", data)

	return result, nil
}

func ExtraString(contents []byte, re *regexp.Regexp) string {
	match := re.FindSubmatch(contents)

	if len(match) >= 2 {
		return string(match[1])
	}

	return ""
}
