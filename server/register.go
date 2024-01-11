package server

import (
	"context"
	"fmt"

	"github.com/sjy-dv/bridger/protobuf/bridgerpb"
	"github.com/sjy-dv/bridger/server/dispatcher"
	"google.golang.org/protobuf/types/known/emptypb"
)

type rpcDispatcher struct {
	bridgerpb.UnimplementedBridgerServer
	DispatchService *dispatchService
}

type dispatchService struct {
	rpcDispatcher
}

func (dispatch *rpcDispatcher) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (dispatch *rpcDispatcher) Dispatch(ctx context.Context, req *bridgerpb.PayloadEmitter) (*bridgerpb.PayloadReceiver, error) {
	type reply struct {
		Result *bridgerpb.PayloadReceiver
		Error  error
	}
	c := make(chan reply, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- reply{
					Result: nil,
					Error:  fmt.Errorf("panic: %v", r),
				}
			}
		}()
		dtx := dispatcher.DispatchContext{}
		dtx.Paylod = req.GetPayload()
		dtx.Domain = req.GetDomain()
		dtx.ExtractMD(ctx)
		c <- reply{
			Result: dispatcher.MatchRoutes(dtx).PayloadReceiver,
			Error:  nil,
		}
	}()
	res := <-c
	return res.Result, res.Error
}
