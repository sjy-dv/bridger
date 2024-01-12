package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/sjy-dv/bridger/client/options"
	pb "github.com/sjy-dv/bridger/grpc/protocol/v0"
	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/keepalive"
)

type BridgerAgent struct {
	*gRpcClientPool
	deadline *time.Duration
}

var reuseOpts = []grpc.DialOption{}

func RegisterBridgerClient(opt *options.Options) *BridgerAgent {
	localLogging := logrus.New()
	if opt.MinChannelSize > opt.MaxChannelSize {
		panic("min channel size can't exceed max channel size")
	}
	maxRecvMsgSize := func() int {
		if opt.MaxRecvMsgSize > 0 {
			localLogging.WithField("action", fmt.Sprintf("grpc_configure_max_recv_msg_size : %v", opt.MaxRecvMsgSize)).
				Info("needs to be synchronized with server")
			return opt.MaxRecvMsgSize
		}
		return options.DefaultMsgSize
	}()
	maxSendMsgSize := func() int {
		if opt.MaxSendMsgSize > 0 {
			localLogging.WithField("action", fmt.Sprintf("grpc_configure_max_send_msg_size : %v", opt.MaxSendMsgSize)).
				Info("needs to be synchronized with server")
			return opt.MaxSendMsgSize
		}
		return options.DefaultMsgSize
	}()
	bridger := &BridgerAgent{}
	agents := &gRpcClientPool{}
	agents.poolSize = &atomic.Int32{}
	agents.maxpoolsize = func() int {
		if opt.MaxChannelSize == 0 {
			localLogging.WithField("action", "bridger-config-max-channel-size").
				Info("bridger-config-max-channel-size default 4")
			return 4
		}
		localLogging.WithField("action", "bridger-config-max-channel-size").
			Info(fmt.Sprintf("bridger-config-max-channel-size custom size %v",
				opt.MaxChannelSize))
		return opt.MaxChannelSize
	}()
	agents.maxsessions = func() int32 {
		if opt.MaxSession == 0 {
			return 100
		}
		return opt.MaxSession
	}()
	agents.minpoolsize = func() int {
		if opt.MinChannelSize == 0 {
			localLogging.WithField("action", "bridger-config-max-channel-size").
				Info("bridger-config-max-channel-size default 1")
			return 1
		}
		localLogging.WithField("action", "bridger-config-max-channel-size").
			Info(fmt.Sprintf("bridger-config-max-channel-size custom size %v",
				opt.MinChannelSize))
		return opt.MinChannelSize
	}()
	agents.addr = opt.Addr
	agents.pool = make([]*connectionpool, opt.MinChannelSize)
	clientOpts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.FailOnNonTempDialError(true),
		grpc.WithReturnConnectionError(),
		grpc.WithDisableRetry(),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  100 * time.Millisecond,
				Multiplier: 1.6,
				Jitter:     0.2,
				MaxDelay:   3 * time.Second,
			},
			MinConnectTimeout: time.Millisecond,
		}),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxRecvMsgSize),
			grpc.MaxCallSendMsgSize(maxSendMsgSize),
		),
	}
	if opt.Credentials {
		clientOpts = append(clientOpts, grpc.WithTransportCredentials(
			credentials.NewTLS(&tls.Config{})))
	} else {
		clientOpts = append(clientOpts, grpc.WithTransportCredentials(
			insecure.NewCredentials()))
	}
	if opt.ClientInterceptor != nil {
		localLogging.WithField("action", "register-interceptor")
		clientOpts = append(clientOpts, grpc.WithUnaryInterceptor(opt.ClientInterceptor))
	}
	if opt.KeepAliveTimeout != 0 && opt.KeepAliveTime != 0 {
		localLogging.WithField("action", "configure-keepalive").
			Info("The keepalive time should be the same as the server.")
		clientOpts = append(clientOpts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                opt.KeepAliveTime,
			Timeout:             opt.KeepAliveTimeout,
			PermitWithoutStream: true,
		}))
	} else {
		clientOpts = append(clientOpts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                options.DefaultKeepAlive,
			Timeout:             options.DefaultKeepAliveTimeout,
			PermitWithoutStream: true,
		}))
	}
	for i := 0; i < opt.MinChannelSize; i++ {
		status := true
		dialContext, cancel := context.WithTimeout(context.Background(), options.DialTimeout)
		conn, err := grpc.DialContext(dialContext, agents.addr, clientOpts...)
		if err != nil {
			conn = nil
			status = false
			localLogging.WithField("action", "bridger-established").
				Info(fmt.Sprintf("bridger-connection-%d is not established", i+1))
			cancel()
			return nil
		}
		if conn != nil {
			localLogging.WithField("action", "bridger-established").
				Info(fmt.Sprintf("bridger-connection-%d is established", i+1))
		}
		agents.pool[i] = &connectionpool{
			connection: conn,
			lastCall:   time.Now(),
			status:     status,
			sessions:   &atomic.Int32{},
		}
		cancel()
	}
	reuseOpts = clientOpts
	bridger = &BridgerAgent{
		agents,
		&opt.Timeout,
	}
	return bridger
}

func (agent *BridgerAgent) Dispatch(domain string, v interface{}, callOptions ...CallOptions) ([]byte, error) {
	data, err := marshal(v)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	if len(callOptions) != 0 {
		if callOptions[0].MetadataHeader != nil {
			md := callOptions[0].MetadataHeader
			for k, v := range md {
				ctx = appendMetaData(ctx, k, v)
			}
		}
	}
	var cancel context.CancelFunc
	if agent.deadline != nil {
		if len(callOptions) > 0 && callOptions[0].Ctx != nil {
			ctx = callOptions[0].Ctx
		} else {
			ctx, cancel = context.WithTimeout(ctx, *agent.deadline)
			defer cancel()
		}
	}

	cp := agent.getPool()
	if cp == nil {
		return nil, errors.New("connection is not connected")
	}
	defer agent.rollbackConnection(cp)
	val, err := pb.NewBridgerClient(cp.connection).Dispatch(
		ctx,
		&pb.PayloadEmitter{
			Payload: data,
			Domain:  domain,
		},
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if val.Info != nil {
		return nil, errors.New(fmt.Sprintf("%s Error: %s", val.Info.Domain, val.Info.Reason))
	}
	return val.GetPayload(), nil
}

func (agents *BridgerAgent) Close() {
	for _, agent := range agents.pool {
		agent.connection.Close()
	}
}

func marshal(v interface{}) ([]byte, error) {
	b, err := msgpack.Marshal(v)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func Unmarshal(b []byte, v interface{}) error {
	err := msgpack.Unmarshal(b, v)
	if err != nil {
		return err
	}
	return nil
}
