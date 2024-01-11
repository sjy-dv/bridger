package server

import (
	"context"
	"fmt"

	pb "github.com/sjy-dv/bridger/grpc/protocol/v0"

	"github.com/sjy-dv/bridger/server/dispatcher"
	"google.golang.org/protobuf/types/known/emptypb"
)

type rpcDispatcher struct {
	pb.UnimplementedBridgerServer
	DispatchService *dispatchService
}

type dispatchService struct {
	rpcDispatcher
}

func (dispatch *rpcDispatcher) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (dispatch *rpcDispatcher) Dispatch(ctx context.Context, req *pb.PayloadEmitter) (*pb.PayloadReceiver, error) {
	type reply struct {
		Result *pb.PayloadReceiver
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
