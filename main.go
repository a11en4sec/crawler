package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/a11en4sec/crawler/storage/sqlstorage"

	"github.com/go-micro/plugins/v4/config/encoder/toml"
	"go-micro.dev/v4/config"
	"go-micro.dev/v4/config/reader"
	"go-micro.dev/v4/config/reader/json"
	"go-micro.dev/v4/config/source"
	"go-micro.dev/v4/config/source/file"

	"go.uber.org/zap"

	"github.com/a11en4sec/crawler/spider"

	"github.com/a11en4sec/crawler/collect"
	"github.com/a11en4sec/crawler/engine"
	"github.com/a11en4sec/crawler/limiter"
	"github.com/a11en4sec/crawler/log"
	"github.com/a11en4sec/crawler/proxy"

	//_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap/zapcore"
	"golang.org/x/time/rate"
)

// Version information.
var (
	BuildTS   = "None"
	GitHash   = "None"
	GitBranch = "None"
	Version   = "None"
)

var (
	PrintVersion = flag.Bool("version", false, "print the version of this build")
)

func main() {
	flag.Parse()

	if *PrintVersion {
		Printer()
		os.Exit(0)
	}

	var (
		err     error
		logger  *zap.Logger
		p       proxy.Func
		storage spider.Storage
	)

	// load config
	enc := toml.NewEncoder()
	cfg, err := config.NewConfig(config.WithReader(json.NewReader(reader.WithEncoder(enc))))

	if err != nil {
		panic(err)
	}

	err = cfg.Load(file.NewSource(
		file.WithPath("config.toml"),
		source.WithEncoder(enc),
	))

	if err != nil {
		panic(err)
	}

	// log
	logText := cfg.Get("logLevel").String("INFO")
	logLevel, err := zapcore.ParseLevel(logText)

	if err != nil {
		panic(err)
	}

	plugin := log.NewStdoutPlugin(logLevel)
	logger = log.NewLogger(plugin)
	logger.Info("log init end")

	// set zap global logger
	zap.ReplaceGlobals(logger)

	// fetch
	proxyURLs := cfg.Get("fetcher", "proxy").StringSlice([]string{})
	timeout := cfg.Get("fetcher", "timeout").Int(5000)
	logger.Sugar().Info("proxy list: ", proxyURLs, " timeout: ", timeout)

	if p, err = proxy.RoundRobinProxySwitcher(proxyURLs...); err != nil {
		logger.Error("RoundRobinProxySwitcher failed", zap.Error(err))
	}

	_ = p

	var f spider.Fetcher = &collect.BrowserFetch{
		Timeout: time.Duration(timeout) * time.Millisecond,
		Logger:  logger,
		//Proxy:   p,
	}

	// storage
	sqlURL := cfg.Get("storage", "sqlURL").String("")
	if storage, err = sqlstorage.New(
		sqlstorage.WithSQLURL(sqlURL),
		sqlstorage.WithLogger(logger.Named("sqlDB")),
		sqlstorage.WithBatchCount(2),
	); err != nil {
		logger.Error("create sqlStorage failed", zap.Error(err))

		return
	}

	// init tasks
	var tcfg []spider.TaskConfig

	if err := cfg.Get("Tasks").Scan(&tcfg); err != nil {
		logger.Error("init seed tasks", zap.Error(err))
	}

	seeds := ParseTaskConfig(logger, f, storage, tcfg)

	s := engine.NewEngine(
		engine.WithFetcher(f),
		engine.WithLogger(logger),
		engine.WithWorkCount(5),
		engine.WithSeeds(seeds),
		engine.WithScheduler(engine.NewSchedule()),
	)

	// start worker
	s.Run()
}

func ParseTaskConfig(logger *zap.Logger, f spider.Fetcher, s spider.Storage, cfgs []spider.TaskConfig) []*spider.Task {
	tasks := make([]*spider.Task, 0, 1000)

	for _, cfg := range cfgs {
		t := spider.NewTask(
			spider.WithName(cfg.Name),
			spider.WithReload(cfg.Reload),
			spider.WithCookie(cfg.Cookie),
			spider.WithLogger(logger),
			spider.WithStorage(s),
		)

		if cfg.WaitTime > 0 {
			t.WaitTime = cfg.WaitTime
		}

		if cfg.MaxDepth > 0 {
			t.MaxDepth = cfg.MaxDepth
		}

		var limits []limiter.RateLimiter

		if len(cfg.Limits) > 0 {
			for _, lcfg := range cfg.Limits {
				// speed limiter
				l := rate.NewLimiter(limiter.Per(lcfg.EventCount, time.Duration(lcfg.EventDur)*time.Second), 1)
				limits = append(limits, l)
			}

			multiLimiter := limiter.Multi(limits...)
			t.Limit = multiLimiter
		}

		switch cfg.Fetcher {
		case "browser":
			t.Fetcher = f
		}

		tasks = append(tasks, t)
	}

	return tasks
}

func GetVersion() string {
	if GitHash != "" {
		h := GitHash
		if len(h) > 7 {
			h = h[:7]
		}

		return fmt.Sprintf("%s-%s", Version, h)
	}

	return Version
}

// Printer print build version
//nolint
func Printer() {
	fmt.Println("Version:          ", GetVersion())
	fmt.Println("Git Branch:       ", GitBranch)
	fmt.Println("Git Commit:       ", GitHash)
	fmt.Println("Build Time (UTC): ", BuildTS)
}
