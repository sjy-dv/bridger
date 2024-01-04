package server

import (
	"errors"
	"fmt"
	"net"

	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/sirupsen/logrus"
	"github.com/sjy-dv/bridger/protobuf/bridgerpb"
	"github.com/sjy-dv/bridger/server/dispatcher"
	"github.com/sjy-dv/bridger/server/options"
	"google.golang.org/grpc"
)

type bridger struct {
	*dispatcher.DispatchAPI
}

func New() *bridger {
	dispatcher.DMap = make(map[string]func(ctx dispatcher.DispatchContext) *dispatcher.ResponseWriter)
	return &bridger{
		&dispatcher.DispatchAPI{},
	}
}

func (b *bridger) RegisterBridgerServer(opt *options.Options) error {
	if opt.Port == 0 {
		return errors.New("port must be specified")
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", opt.Port))
	if err != nil {
		panic(err)
	}
	serverOptions := []grpc.ServerOption{}
	if opt.ChainStreamInterceptorLogger || opt.ChainUnaryInterceptorLogger {
		logrus.ErrorKey = "grpc.error"
		logrusEntry := logrus.NewEntry(logrus.StandardLogger())
		if opt.ChainStreamInterceptorLogger {
			serverOptions = append(serverOptions, grpc.ChainStreamInterceptor(
				grpc_recovery.StreamServerInterceptor(),
				grpc_logrus.StreamServerInterceptor(logrusEntry),
			))
		}
		if opt.ChainUnaryInterceptorLogger {
			serverOptions = append(serverOptions, grpc.ChainUnaryInterceptor(
				grpc_recovery.UnaryServerInterceptor(),
				grpc_logrus.UnaryServerInterceptor(logrusEntry),
			))
		}
	}
	dispatch := rpcDispatcher{}
	dispatch.DispatchService = &dispatchService{dispatch}
	grpcServer := grpc.NewServer(serverOptions...)
	bridgerpb.RegisterBridgerServer(grpcServer, dispatch.DispatchService)
	fmt.Println(fmt.Sprintf("Bridger Server Started on %d", opt.Port))
	if err := grpcServer.Serve(lis); err != nil {
		panic(err)
	}

	return nil
}
