package main

import (
	"github.com/a11en4sec/crawler/collect"
	"github.com/a11en4sec/crawler/engine"
	"github.com/a11en4sec/crawler/limiter"
	"github.com/a11en4sec/crawler/log"
	"github.com/a11en4sec/crawler/proxy"
	"github.com/a11en4sec/crawler/storage"
	"github.com/a11en4sec/crawler/storage/sqlstorage"
	"go.uber.org/zap/zapcore"
	"golang.org/x/time/rate"
	"time"
)

func main() {

	// log
	plugin := log.NewStdoutPlugin(zapcore.InfoLevel)
	logger := log.NewLogger(plugin)
	//logger.Info("log init end")

	// proxy
	proxyURLs := []string{"http://127.0.0.1:8889", "http://127.0.0.1:8889"}
	_, err := proxy.RoundRobinProxySwitcher(proxyURLs...)
	if err != nil {
		logger.Error("RoundRobinProxySwitcher failed")
		return
	}

	// fetcher
	var f collect.Fetcher = &collect.BrowserFetch{
		Timeout: 3000 * time.Millisecond,
		Logger:  logger,
		//Proxy:   p,
	}

	// storage
	var storage storage.Storage
	if storage, err = sqlstorage.New(
		sqlstorage.WithSqlUrl("root:root@r00t@tcp(127.0.0.1:3306)/crawler?charset=utf8"),
		sqlstorage.WithLogger(logger.Named("sqlDB")),
		sqlstorage.WithBatchCount(2),
	); err != nil {
		logger.Error("create sqlstorage failed")
		return
	}

	// limit限流器(多重)
	// 2秒钟1个
	secondLimit := rate.NewLimiter(limiter.Per(1, 2*time.Second), 1)
	// 1分钟20个
	minuteLimit := rate.NewLimiter(limiter.Per(20, 1*time.Minute), 20)
	multiLimiter := limiter.MultiLimiter(secondLimit, minuteLimit)

	// seeds
	var seeds = make([]*collect.Task, 0, 1000)
	seeds = append(seeds, &collect.Task{
		//Name:    "find_douban_sun_room",
		Property: collect.Property{
			Name: "douban_book_list",
		},
		Fetcher: f,
		Storage: storage,
		Limit:   multiLimiter,
	})

	s := engine.NewEngine(
		engine.WithFetcher(f),
		engine.WithLogger(logger),
		engine.WithWorkCount(5),
		engine.WithSeeds(seeds),
		engine.WithScheduler(engine.NewSchedule()),
	)

	// start worker
	go s.Run()

}
