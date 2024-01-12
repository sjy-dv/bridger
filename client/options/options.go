package options

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

type Options struct {
	Addr              string
	MinChannelSize    int
	MaxChannelSize    int
	Timeout           time.Duration
	MaxRecvMsgSize    int
	MaxSendMsgSize    int
	ClientInterceptor func(ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error
	KeepAliveTimeout time.Duration
	KeepAliveTime    time.Duration
	MaxSession       int32
}

const DefaultMsgSize = 104858000 // 10mb
const DialTimeout = 60 * time.Second
const DefaultKeepAliveTimeout = 60 * time.Second
const DefaultKeepAlive = 60 * time.Second
const DefaultMaxSession = 100
