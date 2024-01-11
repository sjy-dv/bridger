// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.19.6
// source: bridger/bridger.proto

package protocol

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Bridger_Ping_FullMethodName     = "/bridger.v0.Bridger/Ping"
	Bridger_Dispatch_FullMethodName = "/bridger.v0.Bridger/Dispatch"
)

// BridgerClient is the client API for Bridger service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BridgerClient interface {
	Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
	Dispatch(ctx context.Context, in *PayloadEmitter, opts ...grpc.CallOption) (*PayloadReceiver, error)
}

type bridgerClient struct {
	cc grpc.ClientConnInterface
}

func NewBridgerClient(cc grpc.ClientConnInterface) BridgerClient {
	return &bridgerClient{cc}
}

func (c *bridgerClient) Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Bridger_Ping_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bridgerClient) Dispatch(ctx context.Context, in *PayloadEmitter, opts ...grpc.CallOption) (*PayloadReceiver, error) {
	out := new(PayloadReceiver)
	err := c.cc.Invoke(ctx, Bridger_Dispatch_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BridgerServer is the server API for Bridger service.
// All implementations should embed UnimplementedBridgerServer
// for forward compatibility
type BridgerServer interface {
	Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	Dispatch(context.Context, *PayloadEmitter) (*PayloadReceiver, error)
}

// UnimplementedBridgerServer should be embedded to have forward compatible implementations.
type UnimplementedBridgerServer struct {
}

func (UnimplementedBridgerServer) Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedBridgerServer) Dispatch(context.Context, *PayloadEmitter) (*PayloadReceiver, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Dispatch not implemented")
}

// UnsafeBridgerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BridgerServer will
// result in compilation errors.
type UnsafeBridgerServer interface {
	mustEmbedUnimplementedBridgerServer()
}

func RegisterBridgerServer(s grpc.ServiceRegistrar, srv BridgerServer) {
	s.RegisterService(&Bridger_ServiceDesc, srv)
}

func _Bridger_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BridgerServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Bridger_Ping_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BridgerServer).Ping(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Bridger_Dispatch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PayloadEmitter)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BridgerServer).Dispatch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Bridger_Dispatch_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BridgerServer).Dispatch(ctx, req.(*PayloadEmitter))
	}
	return interceptor(ctx, in, info, handler)
}

// Bridger_ServiceDesc is the grpc.ServiceDesc for Bridger service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Bridger_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "bridger.v0.Bridger",
	HandlerType: (*BridgerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _Bridger_Ping_Handler,
		},
		{
			MethodName: "Dispatch",
			Handler:    _Bridger_Dispatch_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "bridger/bridger.proto",
}