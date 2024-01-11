package main

import (
	"context"
	"log"

	"github.com/sjy-dv/bridger/server"
	"github.com/sjy-dv/bridger/server/dispatcher"
	"github.com/sjy-dv/bridger/server/options"
	"google.golang.org/grpc"
)

func main() {
	bridger := server.New()

	bridger.Register("/greetings", greetings)
	bridger.Register("/greetings/withname",
		greetingsWithHeaderName,
		"is using metadata api")
	bridger.RegisterBridgerServer(&options.Options{
		Port:                         50051,
		ChainUnaryInterceptorLogger:  false,
		ChainStreamInterceptorLogger: false,
		ServerInterceptor:            sampleInterceptor,
	})
}

func sampleInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	log.Printf("Before handling. Method: %s", info.FullMethod)
	resp, err = handler(ctx, req)
	log.Printf("After handling. Method: %s; Error: %v", info.FullMethod, err)
	return resp, err
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
