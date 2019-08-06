package main

import (
	"context"
	"fmt"
	proto "micro_training/proto"

	micro "github.com/micro/go-micro"
	"github.com/micro/go-micro/registry"
	zookeeper "github.com/micro/go-plugins/registry/zookeeper"
)

func main() {
	reg := zookeeper.NewRegistry(func(op *registry.Options) {
		op.Addrs = []string{
			"118.24.4.114:2181",
			"118.24.4.114:2182",
			"118.24.4.114:2183",
		}
	})

	service := micro.NewService(micro.Name("greeter-client"), micro.Registry(reg))
	service.Init()

	greeter := proto.NewGreeterService("greeter", service.Client())
	rsp, err := greeter.Hello(context.TODO(), &proto.HelloRequest{Name: "John"})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(rsp.GetGreeting())
}
