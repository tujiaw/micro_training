package main

import (
	"context"
	"fmt"
	proto "micro_training/proto"

	micro "github.com/micro/go-micro"
)

type Greeter struct{}

func (g *Greeter) Hello(ctx context.Context, req *proto.HelloRequest, rsp *proto.HelloResponse) error {
	rsp.Greeting = "hello, " + req.Name
	return nil
}

func main() {
	service := micro.NewService(micro.Name("greeter"))
	service.Init()
	proto.RegisterGreeterHandler(service.Server(), new(Greeter))
	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}

// go run main.go --registry=mdns --server_address=localhost:9090
