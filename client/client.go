package client

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/sjy-dv/bridger/client/options"
	"github.com/sjy-dv/bridger/protobuf/bridgerpb"
	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	_ "google.golang.org/grpc/encoding/gzip"
)

type bridgerAgent struct {
	*gRpcClientPool
	deadline *time.Duration
}

func RegisterBridgerClient(opt *options.Options) *bridgerAgent {
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
	bridger := &bridgerAgent{}
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
	for i := 0; i < opt.MinChannelSize; i++ {
		status := true
		conn, err := grpc.Dial(agents.addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(maxRecvMsgSize),
				grpc.MaxCallSendMsgSize(maxSendMsgSize),
			))
		if err != nil {
			conn = nil
			status = false
			localLogging.WithField("action", "bridger-established").
				Info(fmt.Sprintf("bridger-connection-%d is not established", i+1))
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
	}
	bridger = &bridgerAgent{
		agents,
		&opt.Timeout,
	}
	return bridger
}

func (agent *bridgerAgent) Dispatch(domain string, v interface{}, callOPtions ...CallOptions) ([]byte, error) {
	data, err := marshal(v)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	if len(callOPtions) != 0 {
		if callOPtions[0].MetadataHeader != nil {
			md := callOPtions[0].MetadataHeader
			for k, v := range md {
				ctx = appendMetaData(ctx, k, v)
			}
		}
	}
	var cancel context.CancelFunc
	if agent.deadline != nil {
		if len(callOPtions) > 0 && callOPtions[0].Ctx != nil {
			ctx = callOPtions[0].Ctx
		} else {
			ctx, cancel = context.WithTimeout(ctx, *agent.deadline)
			defer cancel()
		}
	}

	val, err := bridgerpb.NewBridgerClient(
		agent.getConnection().connection).Dispatch(
		ctx,
		&bridgerpb.PayloadEmitter{
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

func (agents *bridgerAgent) Close() {
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
