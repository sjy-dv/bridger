package main

import (
	"log"

	"github.com/sjy-dv/bridger/server"
	"github.com/sjy-dv/bridger/server/dispatcher"
	"github.com/sjy-dv/bridger/server/options"
	"google.golang.org/grpc/credentials"
)

func main() {
	bridger := server.New()
	creds, err := credentials.NewServerTLSFromFile("service.pem", "service.key")
	if err != nil {
		log.Fatalf("Failed to setup TLS: %v", err)
	}
	bridger.Register("/greetings", greetings)
	bridger.Register("/greetings/withname",
		greetingsWithHeaderName,
		"is using metadata api")
	bridger.RegisterBridgerServer(&options.Options{
		Port:                         50051,
		ChainUnaryInterceptorLogger:  true,
		ChainStreamInterceptorLogger: true,
		Credentials: options.Credentials{
			Enable: true,
			Cred:   creds,
		},
	})
}

func greetings(dtx dispatcher.DispatchContext) *dispatcher.ResponseWriter {
	var (
		req = struct {
			Msg string
		}{}
		err error
	)
	err = dtx.Bind(&req)
	if err != nil {
		return dtx.Error(err)
	}
	req.Msg = req.Msg + "\n" + "Me too.."
	return dtx.Reply(&req)
}

func greetingsWithHeaderName(dtx dispatcher.DispatchContext) *dispatcher.ResponseWriter {
	var (
		req = struct {
			Msg string
		}{}
		err error
	)
	err = dtx.Bind(&req)
	if err != nil {
		return dtx.Error(err)
	}
	name := dtx.GetMetadata("name")
	req.Msg = "Hello " + name
	return dtx.Reply(&req)
}
