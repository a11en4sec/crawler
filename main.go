package main

import (
	"fmt"
	"github.com/a11en4sec/crawler/collect"
	"github.com/a11en4sec/crawler/log"
	"github.com/a11en4sec/crawler/parse/doubangroup"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"time"
)

func main() {

	// log
	plugin := log.NewStdoutPlugin(zapcore.InfoLevel)
	logger := log.NewLogger(plugin)
	logger.Info("log init end")

	// proxy
	//proxyURLs := []string{"http://127.0.0.1:8889", "http://127.0.0.1:8888"}
	//p, err := proxy.RoundRobinProxySwitcher(proxyURLs...)
	//if err != nil {
	//	logger.Error("RoundRobinProxySwitcher failed")
	//}

	// douban cookie
	//cookie := "bid=-UXUw--yL5g; dbcl2=\"214281202:q0BBm9YC2Yg\"; __yadk_uid=jigAbrEOKiwgbAaLUt0G3yPsvehXcvrs; push_noty_num=0; push_doumail_num=0; __utmz=30149280.1665849857.1.1.utmcsr=accounts.douban.com|utmccn=(referral)|utmcmd=referral|utmcct=/; __utmv=30149280.21428; ck=SAvm; _pk_ref.100001.8cb4=%5B%22%22%2C%22%22%2C1665925405%2C%22https%3A%2F%2Faccounts.douban.com%2F%22%5D; _pk_ses.100001.8cb4=*; __utma=30149280.2072705865.1665849857.1665849857.1665925407.2; __utmc=30149280; __utmt=1; __utmb=30149280.23.5.1665925419338; _pk_id.100001.8cb4=fc1581490bf2b70c.1665849856.2.1665925421.1665849856."
	cookie := "bid=RS__CaCDCpo; __utmc=30149280; gr_user_id=0f187965-001a-4477-a142-054dcf8c2885; __gads=ID=ecf1829514e36fee-22e536eeb3d800a4:T=1669881654:RT=1669881654:S=ALNI_MYe6J5ES_9Zv2EbvKlbjr757-MaRA; __gpi=UID=00000b87f776fb47:T=1669881654:RT=1669881654:S=ALNI_MY418AkDkssxFjf8XmWS4eZ930bCg; __utmz=30149280.1670679172.2.2.utmcsr=time.geekbang.org|utmccn=(referral)|utmcmd=referral|utmcct=/column/article/612328; __yadk_uid=UCMMhWuKxCab7lCaQax2qTd6wMX4RtIO; ll=\"118172\"; _pk_ref.100001.8cb4=%5B%22%22%2C%22%22%2C1670722055%2C%22https%3A%2F%2Ftime.geekbang.org%2Fcolumn%2Farticle%2F612328%22%5D; _pk_ses.100001.8cb4=*; __utma=30149280.514834014.1669881653.1670683468.1670722056.4; dbcl2=\"155500819:GRLH4KG5XG8\"; ck=xmVU; push_noty_num=0; push_doumail_num=0; __utmt=1; __utmv=30149280.15550; _pk_id.100001.8cb4=768b23aedab3c687.1670679171.3.1670723559.1670683468.; __utmb=30149280.5.10.1670722056"
	var worklist []*collect.Request
	for i := 50; i <= 50; i += 25 {
		str := fmt.Sprintf("https://www.douban.com/group/szsh/discussion?start=%d", i)
		worklist = append(worklist, &collect.Request{
			Url:       str,
			Cookie:    cookie,
			ParseFunc: doubangroup.ParseURL,
		})
	}

	//url := "https://book.douban.com/subject/1007305/"

	var f collect.Fetcher = collect.BrowserFetch{
		Timeout: 5000 * time.Millisecond,
		//Proxy:   p,
	}

	for len(worklist) > 0 {
		items := worklist
		worklist = nil

		// 爬取数据
		for _, item := range items {
			//fmt.Println("[+] FetchURL:", item.Url)
			body, err := f.Get(item)
			time.Sleep(1 * time.Second)
			if err != nil {
				logger.Error("read content failed",
					zap.Error(err),
				)
				continue
			}

			fmt.Println("body:", string(body))

			// 处理获取的数据
			res := item.ParseFunc(body, item)

			fmt.Printf("res: %v \n", res.Items)

			// 1 打印获取的结果(正则表达式提取)
			for _, item := range res.Items {
				logger.Info("result",
					zap.String("get url:", item.(string)))
			}
			// 2 将新的待爬取的url放入worklist中
			worklist = append(worklist, res.Requesrts...)
		}
	}

}
