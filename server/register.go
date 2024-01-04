package server

import (
	"context"

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
	dtx := dispatcher.DispatchContext{}
	dtx.Paylod = req.GetPayload()
	dtx.Domain = req.GetDomain()
	dtx.ExtractMD(ctx)
	return dispatcher.MatchRoutes(dtx).PayloadReceiver, nil
}
