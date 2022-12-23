package main

import (
	"github.com/a11en4sec/crawler/collect"
	"github.com/a11en4sec/crawler/collector"
	"github.com/a11en4sec/crawler/collector/sqlstorage"
	"github.com/a11en4sec/crawler/engine"
	"github.com/a11en4sec/crawler/log"
	"github.com/a11en4sec/crawler/proxy"
	"go.uber.org/zap/zapcore"
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
	var storage collector.Storage
	storage, err = sqlstorage.New(
		sqlstorage.WithSqlUrl("root:root@r00t@tcp(127.0.0.1:3306)/crawler?charset=utf8"),
		sqlstorage.WithLogger(logger.Named("sqlDB")),
		sqlstorage.WithBatchCount(2),
	)
	if err != nil {
		logger.Error("create sqlstorage failed")
		return
	}

	// seeds
	var seeds = make([]*collect.Task, 0, 1000)
	seeds = append(seeds, &collect.Task{
		//Name:    "find_douban_sun_room",
		Property: collect.Property{
			Name: "douban_book_list",
		},
		Fetcher: f,
		Storage: storage,
	})

	s := engine.NewEngine(
		engine.WithFetcher(f),
		engine.WithLogger(logger),
		engine.WithWorkCount(5),
		engine.WithSeeds(seeds),
		engine.WithScheduler(engine.NewSchedule()),
	)
	s.Run()
}
