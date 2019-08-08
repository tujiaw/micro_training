package main

import (
	"fmt"

	"github.com/micro/mdns"
)

func main() {

	// Make a channel for results and start listening
	entriesCh := make(chan *mdns.ServiceEntry, 8)
	go func() {
		for entry := range entriesCh {
			fmt.Printf("Got new entry: %v\n", entry)
		}
	}()

	// Start the lookup
	err := mdns.Lookup("greeter", entriesCh)
	if err != nil {
		fmt.Println(err)
	}

	close(entriesCh)
}
