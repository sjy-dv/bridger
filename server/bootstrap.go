package server

import (
	"errors"
	"fmt"
	"net"
	"runtime"

	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/keepalive"

	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/sirupsen/logrus"
	pb "github.com/sjy-dv/bridger/grpc/protocol/v0"
	"github.com/sjy-dv/bridger/server/dispatcher"
	"github.com/sjy-dv/bridger/server/options"
	"google.golang.org/grpc"
)

type bridger struct {
	*dispatcher.DispatchAPI
}

func New() *bridger {
	dispatcher.DMap = make(map[string]func(ctx dispatcher.DispatchContext) *dispatcher.ResponseWriter)
	api := &dispatcher.DispatchAPI{}
	return &bridger{
		api.NewDispatch(),
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
		b.Logger.WithField("action", "grpc_configure_logging")
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
	if opt.MaxRecvMsgSize > 0 || opt.MaxSendMsgSize > 0 {
		if opt.MaxRecvMsgSize > 0 {
			serverOptions = append(serverOptions, grpc.MaxRecvMsgSize(opt.MaxRecvMsgSize))
			b.Logger.WithField("action", fmt.Sprintf("grpc_configure_max_recv_msg_size : %v", opt.MaxRecvMsgSize)).
				Info("needs to be synchronized with clients")
		} else {
			serverOptions = append(serverOptions, grpc.MaxRecvMsgSize(options.DefaultMsgSize))
		}
		if opt.MaxSendMsgSize > 0 {
			serverOptions = append(serverOptions, grpc.MaxSendMsgSize(opt.MaxSendMsgSize))
			b.Logger.WithField("action", fmt.Sprintf("grpc_configure_max_send_msg_size : %v", opt.MaxSendMsgSize)).
				Info("needs to be synchronized with clients")
		} else {
			serverOptions = append(serverOptions, grpc.MaxSendMsgSize(options.DefaultMsgSize))
		}
	} else {
		serverOptions = append(serverOptions, []grpc.ServerOption{
			grpc.MaxRecvMsgSize(options.DefaultMsgSize),
			grpc.MaxSendMsgSize(options.DefaultMsgSize),
		}...)
	}
	if opt.ServerInterceptor != nil {
		b.Logger.WithField("action", "grpc_configure_server_interceptor")
		serverOptions = append(serverOptions, grpc.UnaryInterceptor(opt.ServerInterceptor))
	}
	if opt.KeepAliveTimeout != 0 && opt.KeepAliveTime != 0 {
		serverOptions = append(serverOptions, grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    opt.KeepAliveTime,
			Timeout: opt.KeepAliveTimeout,
		}))
	} else {
		serverOptions = append(serverOptions, grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    options.DefaultKeepAlive,
			Timeout: options.DefaultKeepAliveTimeout,
		}))
	}
	dispatch := rpcDispatcher{}
	dispatch.DispatchService = &dispatchService{dispatch}
	grpcServer := grpc.NewServer(serverOptions...)
	pb.RegisterBridgerServer(grpcServer, dispatch.DispatchService)
	b.Logger.WithField("action", "grpc_startup").Infof("grpc server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		b.Logger.WithError(err)
		runtime.Goexit()
	}
	return nil
}
