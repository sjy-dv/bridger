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
}

const DefaultMsgSize = 104858000 // 10mb
