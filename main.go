package main

import (
	"fmt"
	"github.com/a11en4sec/crawler/collect"
	"github.com/a11en4sec/crawler/engine"
	"github.com/a11en4sec/crawler/log"
	"github.com/a11en4sec/crawler/parse/doubangroup"
	"github.com/a11en4sec/crawler/proxy"
	"go.uber.org/zap/zapcore"
	"time"
)

func main() {

	// log
	plugin := log.NewStdoutPlugin(zapcore.InfoLevel)
	logger := log.NewLogger(plugin)
	logger.Info("log init end")

	// proxy
	proxyURLs := []string{"http://127.0.0.1:8889", "http://127.0.0.1:8889"}
	_, err := proxy.RoundRobinProxySwitcher(proxyURLs...)
	if err != nil {
		logger.Error("RoundRobinProxySwitcher failed")
	}

	// fetcher
	var f collect.Fetcher = &collect.BrowserFetch{
		Timeout: 3000 * time.Millisecond,
		Logger:  logger,
		//Proxy:   p,
	}

	// seeds
	var seeds = make([]*collect.Task, 0, 1000)

	for i := 0; i <= 50; i += 25 {
		str := fmt.Sprintf("https://www.douban.com/group/szsh/discussion?start=%d", i)
		seeds = append(seeds, &collect.Task{
			Url:      str,
			WaitTime: 1 * time.Second,
			MaxDepth: 5,
			Fetcher:  f,
			//Cookie:   "bid=-UXUw--yL5g; dbcl2=\"214281202:q0BBm9YC2Yg\"; __yadk_uid=jigAbrEOKiwgbAaLUt0G3yPsvehXcvrs; push_noty_num=0; push_doumail_num=0; __utmz=30149280.1665849857.1.1.utmcsr=accounts.douban.com|utmccn=(referral)|utmcmd=referral|utmcct=/; __utmv=30149280.21428; ck=SAvm; _pk_ref.100001.8cb4=%5B%22%22%2C%22%22%2C1665925405%2C%22https%3A%2F%2Faccounts.douban.com%2F%22%5D; _pk_ses.100001.8cb4=*; __utma=30149280.2072705865.1665849857.1665849857.1665925407.2; __utmc=30149280; __utmt=1; __utmb=30149280.23.5.1665925419338; _pk_id.100001.8cb4=fc1581490bf2b70c.1665849856.2.1665925421.1665849856.",
			Cookie: "viewed=\"1007305\"; bid=RS__CaCDCpo; __utmc=30149280; gr_user_id=0f187965-001a-4477-a142-054dcf8c2885; __gads=ID=ecf1829514e36fee-22e536eeb3d800a4:T=1669881654:RT=1669881654:S=ALNI_MYe6J5ES_9Zv2EbvKlbjr757-MaRA; __utmz=30149280.1670679172.2.2.utmcsr=time.geekbang.org|utmccn=(referral)|utmcmd=referral|utmcct=/column/article/612328; __yadk_uid=UCMMhWuKxCab7lCaQax2qTd6wMX4RtIO; ll=\"118172\"; dbcl2=\"155500819:GRLH4KG5XG8\"; ck=xmVU; push_noty_num=0; push_doumail_num=0; __utmv=30149280.15550; _vwo_uuid_v2=DFAC5F5F3F11DDE50626BD017B798F6F5|468b57638cbe7a78756b645167bea455; frodotk_db=\"ee6d8efbd50968d9cb48acf37f1c3a8b\"; douban-fav-remind=1; __gpi=UID=00000b87f776fb47:T=1669881654:RT=1670988967:S=ALNI_MY418AkDkssxFjf8XmWS4eZ930bCg; _pk_ref.100001.8cb4=%5B%22%22%2C%22%22%2C1671274367%2C%22https%3A%2F%2Ftime.geekbang.org%2Fcolumn%2Farticle%2F612328%22%5D; _pk_id.100001.8cb4=768b23aedab3c687.1670679171.7.1671274367.1670989663.; _pk_ses.100001.8cb4=*; ap_v=0,6.0; __utma=30149280.514834014.1669881653.1670988968.1671274370.8; __utmt=1; __utmb=30149280.7.5.1671274370",
			RootReq: &collect.Request{
				ParseFunc: doubangroup.ParseURL,
			},
		})
	}

	s := engine.NewEngine(
		engine.WithFetcher(f),
		engine.WithLogger(logger),
		engine.WithWorkCount(5),
		engine.WithSeeds(seeds),
		engine.WithScheduler(engine.NewSchedule()),
	)
	s.Run()
}
