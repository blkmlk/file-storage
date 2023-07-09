package mocks

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"

	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc"

	"github.com/google/uuid"

	"github.com/blkmlk/file-storage/protocol"
)

type FilePart struct {
	ID   string
	Size int64
	Data bytes.Buffer
}

func NewStorage(ctx context.Context) *Storage {
	return &Storage{
		Ctx:       ctx,
		fileParts: make(map[string]*FilePart),
	}
}

type Storage struct {
	Ctx  context.Context
	Size int64

	locker    sync.RWMutex
	fileParts map[string]*FilePart
}

func (s *Storage) GetFileParts() []FilePart {
	s.locker.RLock()
	defer s.locker.RUnlock()

	result := make([]FilePart, 0, len(s.fileParts))
	for _, fp := range s.fileParts {
		result = append(result, *fp)
	}
	return result
}

func (s *Storage) CheckReadiness(ctx context.Context, in *protocol.CheckReadinessRequest, opts ...grpc.CallOption) (*protocol.CheckReadinessResponse, error) {
	if s.Size > 0 && in.Size > s.Size {
		return &protocol.CheckReadinessResponse{Ready: false}, nil
	}
	s.locker.Lock()
	defer s.locker.Unlock()

	id := uuid.NewString()
	s.fileParts[id] = &FilePart{
		ID:   id,
		Size: in.Size,
	}
	return &protocol.CheckReadinessResponse{Id: id, Ready: true}, nil
}

func (s *Storage) CheckFilePartExistence(ctx context.Context, in *protocol.CheckFilePartExistenceRequest, opts ...grpc.CallOption) (*protocol.CheckFilePartExistenceResponse, error) {
	s.locker.RLock()
	defer s.locker.RUnlock()

	_, ok := s.fileParts[in.Id]
	return &protocol.CheckFilePartExistenceResponse{Exists: ok}, nil
}

func (s *Storage) UploadFile(ctx context.Context, opts ...grpc.CallOption) (protocol.Storage_UploadFileClient, error) {
	return &storageUploadStream{
		locker:    &s.locker,
		fileParts: s.fileParts,
	}, nil
}

func (s *Storage) GetFile(ctx context.Context, in *protocol.GetFileRequest, opts ...grpc.CallOption) (protocol.Storage_GetFileClient, error) {
	//TODO implement me
	panic("implement me")
}

type storageUploadStream struct {
	lastID    string
	locker    *sync.RWMutex
	fileParts map[string]*FilePart
}

func (s *storageUploadStream) Send(request *protocol.UploadFileRequest) error {
	s.locker.Lock()
	defer s.locker.Unlock()

	fp, ok := s.fileParts[request.Id]
	if !ok {
		return fmt.Errorf("not found")
	}
	s.lastID = request.Id
	fp.Data.Write(request.Data)

	return nil
}

func (s *storageUploadStream) CloseAndRecv() (*protocol.UploadFileResponse, error) {
	if s.lastID == "" {
		return nil, fmt.Errorf("not found")
	}

	s.locker.Lock()
	defer s.locker.Unlock()

	fp, ok := s.fileParts[s.lastID]
	if !ok {
		return nil, fmt.Errorf("not found")
	}

	h := sha256.New()
	h.Write(fp.Data.Bytes())

	return &protocol.UploadFileResponse{
		Hash: hex.EncodeToString(h.Sum(nil)),
		Size: fp.Size,
	}, nil
}

func (s *storageUploadStream) Header() (metadata.MD, error) {
	//TODO implement me
	panic("implement me")
}

func (s *storageUploadStream) Trailer() metadata.MD {
	//TODO implement me
	panic("implement me")
}

func (s *storageUploadStream) CloseSend() error {
	//TODO implement me
	panic("implement me")
}

func (s *storageUploadStream) Context() context.Context {
	//TODO implement me
	panic("implement me")
}

func (s *storageUploadStream) SendMsg(m interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (s *storageUploadStream) RecvMsg(m interface{}) error {
	//TODO implement me
	panic("implement me")
}
