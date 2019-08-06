package main

import (
	"context"
	"fmt"
	proto "micro_training/proto"

	micro "github.com/micro/go-micro"
	"github.com/micro/go-micro/registry"
	zookeeper "github.com/micro/go-plugins/registry/zookeeper"
)

type Greeter struct{}

func (g *Greeter) Hello(ctx context.Context, req *proto.HelloRequest, rsp *proto.HelloResponse) error {
	rsp.Greeting = "hello, " + req.Name
	return nil
}

func main() {
	reg := zookeeper.NewRegistry(func(op *registry.Options) {
		op.Addrs = []string{
			"118.24.4.114:2181",
			"118.24.4.114:2182",
			"118.24.4.114:2183",
		}
	})

	service := micro.NewService(micro.Name("greeter"), micro.Registry(reg))
	service.Init()
	proto.RegisterGreeterHandler(service.Server(), new(Greeter))
	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}
