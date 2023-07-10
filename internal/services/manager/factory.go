package manager

import (
	"context"

	"github.com/blkmlk/file-storage/internal/mocks"

	"google.golang.org/grpc"

	"github.com/blkmlk/file-storage/protocol"
)

type ClientFactory interface {
	NewStorageClient(ctx context.Context, host string) (protocol.StorageClient, error)
}

func NewGRPCClientFactory() ClientFactory {
	return grpcClientFactory{}
}

func NewMockedClientFactory() ClientFactory {
	return mockedClientFactory{}
}

type grpcClientFactory struct {
}

func (g grpcClientFactory) NewStorageClient(ctx context.Context, host string) (protocol.StorageClient, error) {
	conn, err := grpc.DialContext(ctx, host, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return protocol.NewStorageClient(conn), nil
}

type mockedClientFactory struct {
}

func (m mockedClientFactory) NewStorageClient(ctx context.Context, host string) (protocol.StorageClient, error) {
	return mocks.NewStorage(ctx), nil
}
