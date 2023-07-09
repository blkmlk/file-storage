package mocks

import (
	"context"

	"github.com/blkmlk/file-storage/protocol"
	"google.golang.org/grpc"
)

type Storage struct {
	Ctx context.Context
}

func (s *Storage) CheckReadiness(ctx context.Context, in *protocol.CheckReadinessRequest, opts ...grpc.CallOption) (*protocol.CheckReadinessResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Storage) CheckFilePartExistence(ctx context.Context, in *protocol.CheckFilePartExistenceRequest, opts ...grpc.CallOption) (*protocol.CheckFilePartExistenceResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Storage) UploadFile(ctx context.Context, opts ...grpc.CallOption) (protocol.Storage_UploadFileClient, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Storage) GetFile(ctx context.Context, in *protocol.GetFileRequest, opts ...grpc.CallOption) (protocol.Storage_GetFileClient, error) {
	//TODO implement me
	panic("implement me")
}
