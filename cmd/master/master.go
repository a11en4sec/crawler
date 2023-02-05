package master

import (
	"context"
	"fmt"
	"net/http"
	"time"

	grpccli "github.com/go-micro/plugins/v4/client/grpc"

	//"github.com/a11en4sec/crawler/proto/crawler"
	proto "github.com/a11en4sec/crawler/proto/crawler"

	"github.com/a11en4sec/crawler/cmd/worker"
	"github.com/a11en4sec/crawler/spider"

	"github.com/a11en4sec/crawler/master"

	"github.com/go-micro/plugins/v4/registry/etcd"

	"github.com/spf13/cobra"

	"github.com/a11en4sec/crawler/log"
	"github.com/a11en4sec/crawler/proto/greeter"
	"github.com/go-micro/plugins/v4/config/encoder/toml"
	"github.com/go-micro/plugins/v4/server/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/config"
	"go-micro.dev/v4/config/reader"
	"go-micro.dev/v4/config/reader/json"
	"go-micro.dev/v4/config/source"
	"go-micro.dev/v4/config/source/file"
	"go-micro.dev/v4/registry"
	"go-micro.dev/v4/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	grpc2 "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var MasterCmd = &cobra.Command{
	Use:   "master",
	Short: "run master service.",
	Long:  "run master service.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		Run()
	},
}

var PProfListenAddress string
var masterID string
var HTTPListenAddress string
var GRPCListenAddress string

func init() {
	// 给子命令master 设置flag，./main master [--pprof=9981 | --id=1 | --http=8081 | --grpc=9091]
	MasterCmd.Flags().StringVar(
		&PProfListenAddress, "pprof", ":9981", "set pprof address")
	MasterCmd.Flags().StringVar(
		&masterID, "id", "1", "set master id")
	MasterCmd.Flags().StringVar(
		&HTTPListenAddress, "http", ":8081", "set HTTP listen address")
	MasterCmd.Flags().StringVar(
		&GRPCListenAddress, "grpc", ":9091", "set GRPC listen address")

}

func Run() {
	// start pprof
	go func() {
		if err := http.ListenAndServe(PProfListenAddress, nil); err != nil {
			panic(err)
		}
	}()

	var (
		err    error
		logger *zap.Logger
	)

	// load config
	enc := toml.NewEncoder()
	cfg, err := config.NewConfig(config.WithReader(json.NewReader(reader.WithEncoder(enc))))
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

	//
	fmt.Println("hello master")

	var sconfig ServerConfig
	if err := cfg.Get("MasterServer").Scan(&sconfig); err != nil {
		logger.Error("get GRPC Server config failed", zap.Error(err))
	}
	logger.Sugar().Debugf("grpc server config,%+v", sconfig)

	reg := etcd.NewRegistry(registry.Addrs(sconfig.RegistryAddress))

	// init tasks
	var tcfg []spider.TaskConfig
	if err := cfg.Get("Tasks").Scan(&tcfg); err != nil {
		logger.Error("init seed tasks", zap.Error(err))
	}
	seeds := worker.ParseTaskConfig(logger, nil, nil, tcfg)

	m, err := master.New(
		masterID,
		master.WithLogger(logger.Named("master")),
		master.WithGRPCAddress(GRPCListenAddress),
		master.WithRegistryUrl(sconfig.RegistryAddress),
		master.WithRegistry(reg), // 将etcd注入到go-micro中
		master.WithSeeds(seeds),
	)

	_ = m

	if err != nil {
		logger.Error("init  master falied", zap.Error(err))
	}

	// start http proxy to GRPC
	go RunHTTPServer(sconfig)

	// start grpc server
	RunGRPCServer(m, logger, reg, sconfig)
}

type ServerConfig struct {
	//GRPCListenAddress string
	//HTTPListenAddress string
	//ID               string
	RegistryAddress  string
	RegisterTTL      int
	RegisterInterval int
	Name             string
	ClientTimeOut    int
}

func RunGRPCServer(m *master.Master, logger *zap.Logger, reg registry.Registry, cfg ServerConfig) {
	//reg := etcd.NewRegistry(registry.Addrs(cfg.RegistryAddress))
	service := micro.NewService(
		micro.Server(grpc.NewServer(
			server.Id(masterID),
		)),
		micro.Address(GRPCListenAddress),
		micro.Registry(reg),
		micro.RegisterTTL(time.Duration(cfg.RegisterTTL)*time.Second),
		micro.RegisterInterval(time.Duration(cfg.RegisterInterval)*time.Second),
		micro.WrapHandler(logWrapper(logger)),
		micro.Name(cfg.Name),
		micro.Client(grpccli.NewClient()),
	)

	cl := proto.NewCrawlerMasterService(cfg.Name, service.Client())

	m.SetForwardCli(cl)

	// 设置micro 客户端默认超时时间为10秒钟
	if err := service.Client().Init(client.RequestTimeout(time.Duration(cfg.ClientTimeOut) * time.Second)); err != nil {
		logger.Sugar().Error("micro client init error. ", zap.String("error:", err.Error()))

		return
	}

	service.Init()

	if err := greeter.RegisterGreeterHandler(service.Server(), new(Greeter)); err != nil {
		logger.Fatal("register handler failed", zap.Error(err))
	}

	// 注册为grpc服务2
	if err := proto.RegisterCrawlerMasterHandler(service.Server(), m); err != nil {
		logger.Fatal("register handler failed", zap.Error(err))
	}
	if err := service.Run(); err != nil {
		logger.Fatal("grpc server stop", zap.Error(err))
	}
}

type Greeter struct{}

func (g *Greeter) Hello(ctx context.Context, req *greeter.Request, rsp *greeter.Response) error {
	rsp.Greeting = "Hello " + req.Name

	return nil
}

func RunHTTPServer(cfg ServerConfig) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc2.DialOption{
		grpc2.WithTransportCredentials(insecure.NewCredentials()),
	}

	// http代理调用grpc
	//if err := greeter.RegisterGreeterGwFromEndpoint(ctx, mux, GRPCListenAddress, opts); err != nil {
	//	zap.L().Fatal("Register backend grpc server endpoint failed", zap.Error(err))
	//}

	if err := proto.RegisterCrawlerMasterGwFromEndpoint(ctx, mux, GRPCListenAddress, opts); err != nil {
		zap.L().Fatal("Register backend grpc server endpoint failed", zap.Error(err))
	}

	zap.S().Debugf("start master http server listening on %v proxy to grpc server;%v", HTTPListenAddress, GRPCListenAddress)

	if err := http.ListenAndServe(HTTPListenAddress, mux); err != nil {
		zap.L().Fatal("http listenAndServe failed", zap.Error(err))
	}
}

func logWrapper(log *zap.Logger) server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {

			log.Info("receive request",
				zap.String("method", req.Method()),
				zap.String("Service", req.Service()),
				zap.Reflect("request param:", req.Body()),
			)

			err := fn(ctx, req, rsp)

			return err
		}
	}
}