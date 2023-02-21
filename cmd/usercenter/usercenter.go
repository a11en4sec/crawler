package usercenter

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/a11en4sec/crawler/auth"

	"google.golang.org/grpc/metadata"

	"github.com/a11en4sec/crawler/proto/user"
	"github.com/a11en4sec/crawler/userCenter"

	_ "github.com/go-micro/plugins/v4/auth/jwt"

	grpccli "github.com/go-micro/plugins/v4/client/grpc"
	ratePlugin "github.com/go-micro/plugins/v4/wrapper/ratelimiter/ratelimit"
	"github.com/juju/ratelimit"

	"github.com/go-micro/plugins/v4/wrapper/breaker/hystrix"

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

var (
	name    = "helloworld"
	version = "latest"
)

var UserCenterCmd = &cobra.Command{
	Use:   "usercenter",
	Short: "run uc service.",
	Long:  "run uc service.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		Run()
	},
}

var PProfListenAddress string
var masterID string
var HTTPListenAddress string
var GRPCListenAddress string
var cfgFile string

func init() {
	// 给子命令master 设置flag，./main usercenter [--pprof=9981 | --id=1 | --http=8081 | --grpc=9091]
	UserCenterCmd.Flags().StringVar(
		&PProfListenAddress, "pprof", ":49981", "set pprof address")
	UserCenterCmd.Flags().StringVar(
		&cfgFile, "config", "config.toml", "set config file")
	UserCenterCmd.Flags().StringVar(
		&HTTPListenAddress, "http", ":4080", "set HTTP listen address")
	UserCenterCmd.Flags().StringVar(
		&GRPCListenAddress, "grpc", ":4180", "set GRPC listen address")
	UserCenterCmd.Flags().StringVar(
		&masterID, "id", "1", "set uc id")

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
		file.WithPath(cfgFile),
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
	fmt.Println("[+] User Center Service...")

	var sconfig ServerConfig
	if err := cfg.Get("UserCenter").Scan(&sconfig); err != nil {
		logger.Error("get GRPC Server config failed", zap.Error(err))
	}
	logger.Sugar().Debugf("grpc server config,%+v", sconfig)

	reg := etcd.NewRegistry(registry.Addrs(sconfig.RegistryAddress))

	if err != nil {
		logger.Error("init  master falied", zap.Error(err))
	}

	// start http proxy to GRPC
	go RunHTTPServer(sconfig)

	// start grpc server
	RunGRPCServer(logger, reg, sconfig)
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

func RunGRPCServer(logger *zap.Logger, reg registry.Registry, cfg ServerConfig) {
	//reg := etcd.NewRegistry(registry.Addrs(cfg.RegistryAddress))

	// 令牌桶
	// 第一个参数为 0.5，表示每秒钟放入的令牌个数为 0.5 个
	// 第二个参数为桶的大小
	b := ratelimit.NewBucketWithRate(1, 2)

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
		// micro.WrapHandler包装了go-micro的Server端设置的限流中间件
		// ratePlugin.NewHandlerWrapper 的第一个参数为之前设置的令牌桶，
		// 第二个参数可以指定当请求速率超过阈值时，是否堵塞住。此处为 false，表示不堵塞并立即返回错误
		micro.WrapHandler(ratePlugin.NewHandlerWrapper(b, false)),
		// 熔断器
		micro.WrapClient(hystrix.NewClientWrapper()),
	)

	// 修改所有接口的默认熔断参数
	hystrix.ConfigureDefault(hystrix.CommandConfig{
		Timeout:                10000, // 请求的超时时间
		MaxConcurrentRequests:  100,   // 最大的并发数量
		RequestVolumeThreshold: 10,    // 触发断路器的最小数量（避免小量请求的干扰）
		SleepWindow:            6000,  // 断路器打开状态时，等待多长时间再次检测当前链路的状态
		ErrorPercentThreshold:  30,    // 失败率的阈值，当失败率超过该阈值时，将触发熔断
	})

	// 设置micro 客户端默认超时时间为10秒钟
	if err := service.Client().Init(client.RequestTimeout(time.Duration(cfg.ClientTimeOut) * time.Second)); err != nil {
		logger.Sugar().Error("micro client init error. ", zap.String("error:", err.Error()))

		return
	}

	service.Init(
		micro.Name(name),
		micro.Version(version),
		micro.WrapHandler(auth.NewAuthWrapper(service)),
	)

	if err := user.RegisterUserHandler(service.Server(), new(userCenter.LoginService)); err != nil {
		logger.Fatal("register handler failed", zap.Error(err))
	}

	if err := greeter.RegisterGreeterHandler(service.Server(), new(Greeter)); err != nil {
		logger.Fatal("register handler failed", zap.Error(err))
	}

	if err := service.Run(); err != nil {
		logger.Fatal("grpc server stop", zap.Error(err))
	}
}

type Greeter struct{}

func (g *Greeter) Hello(ctx context.Context, req *greeter.Request, rsp *greeter.Response) error {

	var token string

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if ts, ok := md["authorization"]; ok {
			token = strings.Join(ts, ",")
		}
	}

	myclain, err := userCenter.VerifyToken(token)
	if err != nil {
		return err
	}
	u := myclain.Username
	//id := myclain.UserID
	//token := fmt.Sprintf("%v", ctx.Value("Authorization"))
	rsp.Greeting = "Hello " + req.Name + "from token your are: " + u

	return nil
}

func RunHTTPServer(cfg ServerConfig) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	// 将http header 注入到grpc的metadata中
	mux := runtime.NewServeMux(runtime.WithIncomingHeaderMatcher(CustomMatcher))
	opts := []grpc2.DialOption{
		grpc2.WithTransportCredentials(insecure.NewCredentials()),
	}

	// 1 http代理调用grpc（userCenter）
	if err := user.RegisterUserGwFromEndpoint(ctx, mux, GRPCListenAddress, opts); err != nil {
		zap.L().Fatal("Register backend grpc server endpoint failed", zap.Error(err))
	}

	// 2 http代理调用grpc (greeter/hello)
	if err := greeter.RegisterGreeterGwFromEndpoint(ctx, mux, GRPCListenAddress, opts); err != nil {
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

			fmt.Printf("req.Body:%v \n", req.Body())

			log.Info("receive request",
				zap.String("method", req.Method()),
				zap.String("Service", req.Service()),
				zap.String("Endpoint", req.Endpoint()),
				zap.Reflect("request param:", req.Body()),
			)

			err := fn(ctx, req, rsp)

			return err
		}
	}
}

func CustomMatcher(key string) (string, bool) {
	switch key {
	case "X-User-Id":
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}
