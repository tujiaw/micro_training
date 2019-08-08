package main

import (
	"context"
	"fmt"
	proto "micro_training/proto"

	micro "github.com/micro/go-micro"
)

func main() {
	service := micro.NewService(micro.Name("greeter-client"))
	service.Init()

	greeter := proto.NewGreeterService("greeter", service.Client())
	rsp, err := greeter.Hello(context.TODO(), &proto.HelloRequest{Name: "John"})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(rsp.GetGreeting())
}
