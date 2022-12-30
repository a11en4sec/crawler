package main

import (
	"context"
	pb "github.com/a11en4sec/crawler/test/proto/proto/greeter"
	"google.golang.org/grpc"
	"log"
	"net"
)

type Greeter struct {
	pb.UnimplementedGreeterServer
}

// Hello 是接口GreeterServer的方法，
func (g *Greeter) Hello(ctx context.Context, req *pb.Request) (rsp *pb.Response, err error) {
	rsp.Greeting = "Hello " + req.Name
	return rsp, nil
}

func main() {
	println("gRPC server tutorial in Go")

	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()

	// 将结构体注册进grpc server
	pb.RegisterGreeterServer(s, &Greeter{})

	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
