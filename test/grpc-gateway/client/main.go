package main

import (
	"context"
	"fmt"
	grpccli "github.com/go-micro/plugins/v4/client/grpc"
	etcdReg "github.com/go-micro/plugins/v4/registry/etcd"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	pb "grpc-gateway/proto/greeter"
)

func main() {
	reg := etcdReg.NewRegistry(
		registry.Addrs(":2379"),
	)
	// 选项模式进行配置
	service := micro.NewService(
		micro.Registry(reg),
		micro.Client(grpccli.NewClient()),
	)
	// 初始化配置
	service.Init()

	// Use the generated client stub
	cl := pb.NewGreeterService("go.micro.server.worker", service.Client())

	// Make request
	rsp, err := cl.Hello(context.Background(), &pb.Request{
		Name: "John",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(rsp.Greeting)
}
