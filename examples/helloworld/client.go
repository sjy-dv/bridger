package main

import (
	"log"
	"time"

	"github.com/sjy-dv/bridger/client"
	"github.com/sjy-dv/bridger/client/options"
)

func main() {
	/**
	default value
	if you want to singleton instance,
	min&max channel size should be set 1
	*/
	bridgerClient := client.RegisterBridgerClient(&options.Options{
		Addr:           "127.0.0.1:50051",
		MinChannelSize: 1,
		MaxChannelSize: 4,
		Timeout:        time.Duration(time.Second * 5),
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
