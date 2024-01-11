package client

import (
	"context"

	"google.golang.org/grpc/metadata"
)

type MetadataHeader map[string]string

func appendMetaData(ctx context.Context, key, value string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, key, value)
}

type CallOptions struct {
	MetadataHeader MetadataHeader
	Ctx            context.Context
}
