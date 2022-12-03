package main

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"time"
)

// Logger  性能好，需要指定类型
func Logger() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	url := "www.google.com"
	logger.Info("failed to fetch URL",
		zap.String("url", url),
		zap.Int("attempt", 3),
		zap.Duration("backoff", time.Second))
}

// SugaredLogger 性能没有Logger，打印更灵活，不必指定内容类型
func SugaredLogger() {
	logger, _ := zap.NewProduction(zap.WithCaller(false), zap.AddStacktrace(zap.InfoLevel))
	defer logger.Sync()
	sugar := logger.Sugar()
	url := "www.google.com"
	sugar.Infow("failed to fetch URL",
		"url", url,
		"attempt", 3,
		"backoff", time.Second,
	)
}

func loggerUserDefined() {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.EncoderConfig.TimeKey = "timestamp"
	loggerConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	logger, err := loggerConfig.Build()
	if err != nil {
		log.Fatal(err)
	}

	sugar := logger.Sugar()
	sugar.Info("Hello from zap logger")
}

// lumberjack.v2日志切割组件，实现了zapcore.WriteSyncer接口，可以集成到 Zap 中

func main() {
	Logger()
	SugaredLogger()
	loggerUserDefined()
}
