package options

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

type Options struct {
	Port                         int
	ChainUnaryInterceptorLogger  bool
	ChainStreamInterceptorLogger bool
	MaxRecvMsgSize               int
	MaxSendMsgSize               int
	ServerInterceptor            func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error)
	KeepAliveTimeout         time.Duration
	KeepAliveTime            time.Duration
	EnforcementPolicyMinTime time.Duration
}

const (
	b  = 1
	kb = 1024
	mb = 1024 * 1024
	gb = 1024 * 1024 * 1024

	B  = 1
	KB = 1024
	MB = 1024 * 1024
	GB = 1024 * 1024 * 1024
)

const DefaultMsgSize = 104858000 // 10mb
const DefaultKeepAliveTimeout = 10 * time.Second
const DefaultKeepAlive = 60 * time.Second
const DefaultEnforcementPolicyMinTime = 5 * time.Second
