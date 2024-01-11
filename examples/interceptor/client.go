package main

import (
	"context"
	"log"
	"time"

	"github.com/sjy-dv/bridger/client"
	"github.com/sjy-dv/bridger/client/options"
	"google.golang.org/grpc"
)

func main() {
	/**
	default value
	if you want to singleton instance,
	min&max channel size should be set 1
	*/
	bridgerClient := client.RegisterBridgerClient(&options.Options{
		Addr:              "127.0.0.1:50051",
		MinChannelSize:    1,
		MaxChannelSize:    4,
		Timeout:           time.Duration(time.Second * 5),
		ClientInterceptor: sampleClient,
	})
	defer bridgerClient.Close()
	type req struct {
		Msg string
	}
	val, err := bridgerClient.Dispatch("/greetings", &req{Msg: "Hello, Dispatcher"})
	if err != nil {
		panic(err)
	}
	response := &req{}
	err = client.Unmarshal(val, response)
	if err != nil {
		panic(err)
	}
	log.Println("First Message : ", response.Msg)

	header := client.MetadataHeader{}
	header["name"] = "gopher"
	val, err = bridgerClient.Dispatch("/greetings/withname", &req{Msg: "I'm gopher"}, client.CallOptions{
		MetadataHeader: header,
	})
	if err != nil {
		panic(err)
	}
	response = &req{}
	err = client.Unmarshal(val, response)
	if err != nil {
		panic(err)
	}
	log.Println("Second Message : ", response.Msg)
}

func sampleClient(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	log.Printf("Before calling method: %s", method)
	start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	log.Printf("After calling method: %s; Duration: %s; Error: %v", method, time.Since(start), err)
	return err
}
