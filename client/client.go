package client

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/sjy-dv/bridger/client/options"
	"github.com/sjy-dv/bridger/protobuf/bridgerpb"
	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type bridgerAgent struct {
	*gRpcClientPool
	deadline *time.Duration
}

func RegisterBridgerClient(opt *options.Options) *bridgerAgent {
	if opt.MinChannelSize > opt.MaxChannelSize {
		panic("min channel size can't exceed max channel size")
	}
	bridger := &bridgerAgent{}
	agents := &gRpcClientPool{}
	agents.poolSize = &atomic.Int32{}
	agents.maxpoolsize = func() int {
		if opt.MaxChannelSize == 0 {
			return 4
		}
		return opt.MaxChannelSize
	}()
	agents.minpoolsize = func() int {
		if opt.MinChannelSize == 0 {
			return 1
		}
		return opt.MinChannelSize
	}()
	agents.addr = opt.Addr
	agents.pool = make([]*connectionpool, opt.MinChannelSize)
	for i := 0; i < opt.MinChannelSize; i++ {
		status := true
		conn, err := grpc.Dial(agents.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			conn = nil
			status = false
			panic(fmt.Sprintf("bridger-connection-%d is not established", i))
		}
		if conn != nil {
			log.Println(fmt.Sprintf("bridger-connection-%d is not established", i))
		}
		agents.pool[i] = &connectionpool{
			connection: conn,
			lastCall:   time.Now(),
			status:     status,
			sessions:   &atomic.Int32{},
		}
	}
	bridger = &bridgerAgent{agents, &opt.Timeout}
	return bridger
}

func (agent *bridgerAgent) Dispatch(domain string, v interface{}, metadata ...MetadataHeader) ([]byte, error) {
	data, err := marshal(v)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	if len(metadata) != 0 {
		md := metadata[0]
		for k, v := range md {
			ctx = appendMetaData(ctx, k, v)
		}
	}
	var cancel context.CancelFunc
	if agent.deadline != nil {
		ctx, cancel = context.WithTimeout(ctx, *agent.deadline)
		defer cancel()
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
