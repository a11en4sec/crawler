package doubanbook

import (
	"regexp"
	"strconv"
	"sync"

	"github.com/a11en4sec/crawler/collect"
	"go.uber.org/zap"
)

// DoubanBookTask 引擎初始化的时候加载
var DoubanBookTask = &collect.Task{
	Property: collect.Property{
		Name: "douban_book_list",
		//URL:      "",
		//Cookie:   "bid=-UXUw--yL5g; dbcl2=\\\"214281202:q0BBm9YC2Yg\\\"; __yadk_uid=jigAbrEOKiwgbAaLUt0G3yPsvehXcvrs; push_noty_num=0; push_doumail_num=0; __utmz=30149280.1665849857.1.1.utmcsr=accounts.douban.com|utmccn=(referral)|utmcmd=referral|utmcct=/; __utmv=30149280.21428; ck=SAvm; _pk_ref.100001.8cb4=%5B%22%22%2C%22%22%2C1665925405%2C%22https%3A%2F%2Faccounts.douban.com%2F%22%5D; _pk_ses.100001.8cb4=*; __utma=30149280.2072705865.1665849857.1665849857.1665925407.2; __utmc=30149280; __utmt=1; __utmb=30149280.23.5.1665925419338; _pk_id.100001.8cb4=fc1581490bf2b70c.1665849856.2.1665925421.1665849856.",
		Cookie:   "viewed=\"1007305\"; bid=RS__CaCDCpo; __utmc=30149280; __utmc=81379588; gr_user_id=0f187965-001a-4477-a142-054dcf8c2885; __gads=ID=ecf1829514e36fee-22e536eeb3d800a4:T=1669881654:RT=1669881654:S=ALNI_MYe6J5ES_9Zv2EbvKlbjr757-MaRA; __yadk_uid=V07p4iGoUkd9XtJFt50baFxy8AhjMLUi; dbcl2=\"155500819:GRLH4KG5XG8\"; ck=xmVU; push_noty_num=0; push_doumail_num=0; __utmv=30149280.15550; _vwo_uuid_v2=DFAC5F5F3F11DDE50626BD017B798F6F5|468b57638cbe7a78756b645167bea455; frodotk_db=\"ee6d8efbd50968d9cb48acf37f1c3a8b\"; douban-fav-remind=1; ll=\"118282\"; __utmz=30149280.1671668220.13.3.utmcsr=time.geekbang.org|utmccn=(referral)|utmcmd=referral|utmcct=/column/article/615675; _vwo_uuid_v2=DFAC5F5F3F11DDE50626BD017B798F6F5|468b57638cbe7a78756b645167bea455; __gpi=UID=00000b87f776fb47:T=1669881654:RT=1671679841:S=ALNI_MY418AkDkssxFjf8XmWS4eZ930bCg; __utmz=81379588.1671679884.3.3.utmcsr=douban.com|utmccn=(referral)|utmcmd=referral|utmcct=/; __utma=81379588.369659161.1669881653.1671679884.1671687540.4; _pk_ref.100001.3ac3=%5B%22%22%2C%22%22%2C1671687541%2C%22https%3A%2F%2Fwww.douban.com%2F%22%5D; _pk_id.100001.3ac3=42afaf32977203ee.1669881653.4.1671688743.1671680562.; __utma=30149280.514834014.1669881653.1671687540.1671693097.16",
		WaitTime: 2,
		Reload:   false,
		MaxDepth: 5,
	},
	Visited:     nil,
	VisitedLock: sync.Mutex{},
	Rule: collect.RuleTree{
		Root: func() ([]*collect.Request, error) {
			roots := []*collect.Request{
				{
					//Task:     nil,
					URL:      "https://book.douban.com",
					Method:   "GET",
					Priority: 1,
					Depth:    0,
					RuleName: "数据tag",
				},
			}

			return roots, nil
		},
		Trunk: map[string]*collect.Rule{
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
	Fetcher: nil,
}

const regexpStr = `<a href="([^"]+)" class="tag">([^<]+)</a>`

func ParseTag(ctx *collect.Context) (collect.ParseResult, error) {
	re := regexp.MustCompile(regexpStr)

	matches := re.FindAllSubmatch(ctx.Body, -1)
	result := collect.ParseResult{}

	if len(matches) == 0 {
		zap.L().Info(" regex ", zap.String("regex not match", "tag not match"))
	}

	for _, m := range matches {
		result.Requesrts = append(
			result.Requesrts, &collect.Request{
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

func ParseBookList(ctx *collect.Context) (collect.ParseResult, error) {
	re := regexp.MustCompile(BooklistRe)
	matches := re.FindAllSubmatch(ctx.Body, -1)
	result := collect.ParseResult{}

	for _, m := range matches {
		req := &collect.Request{
			Method:   "GET",
			Task:     ctx.Req.Task,
			URL:      string(m[1]),
			Depth:    ctx.Req.Depth + 1,
			RuleName: "书籍简介",
		}
		// 书名缓存到了临时的 tmp 结构中供下一个阶段读取
		req.TmpData = &collect.Temp{}
		if err := req.TmpData.Set("book_name", string(m[2])); err != nil {
			zap.L().Error("Set TmpData failed", zap.Error(err))
		}

		result.Requesrts = append(result.Requesrts, req)
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

func ParseBookDetail(ctx *collect.Context) (collect.ParseResult, error) {
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

	result := collect.ParseResult{
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
