package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/blkmlk/file-storage/env"

	"github.com/blkmlk/file-storage/internal/services/filestorage"

	"github.com/google/uuid"

	"github.com/blkmlk/file-storage/protocol"
)

type Storage struct {
	id           string
	registryHost string
	storageHost  string
	fileStorage  filestorage.FileStorage

	locker   sync.RWMutex
	prepared map[string]bool

	protocol.StorageServer
}

func New(fileStorage filestorage.FileStorage) (*Storage, error) {
	registryHost, err := env.Get(env.RegistryHost)
	if err != nil {
		return nil, err
	}

	storageHost, err := env.Get(env.StorageHost)
	if err != nil {
		return nil, err
	}

	s := &Storage{
		registryHost: registryHost,
		storageHost:  storageHost,
		fileStorage:  fileStorage,
		prepared:     make(map[string]bool),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancel()

	if err = s.register(ctx); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Storage) register(ctx context.Context) error {
	conn, err := grpc.DialContext(ctx, s.registryHost)
	if err != nil {
		return err
	}

	uploaderClient := protocol.NewUploaderClient(conn)
	_, err = uploaderClient.Register(ctx, &protocol.RegisterRequest{
		StorageId: s.id,
		Host:      s.storageHost,
	})
	return err
}

func (s *Storage) CheckReadiness(ctx context.Context, request *protocol.CheckReadinessRequest) (*protocol.CheckReadinessResponse, error) {
	//todo: check for available space
	id := uuid.NewString()
	s.locker.Lock()
	defer s.locker.Unlock()

	s.prepared[id] = true
	return &protocol.CheckReadinessResponse{Id: id, Ready: true}, nil
}

func (s *Storage) CheckFilePartExistence(ctx context.Context, request *protocol.CheckFilePartExistenceRequest) (*protocol.CheckFilePartExistenceResponse, error) {
	exists, err := s.fileStorage.Exists(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &protocol.CheckFilePartExistenceResponse{
		Exists: exists,
	}, nil
}

func (s *Storage) UploadFile(server protocol.Storage_UploadFileServer) error {
	var writer io.WriteCloser
	h := sha256.New()
	received := 0
	for {
		msg, err := server.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
		}

		if writer == nil {
			writer, err = s.fileStorage.Create(server.Context(), msg.Id)
			if err != nil {
				return err
			}
		}
		n, err := writer.Write(msg.Data)
		if err != nil {
			return err
		}
		h.Write(msg.Data)
		received += n
	}

	if writer == nil {
		return fmt.Errorf("no data is written")
	}

	if err := writer.Close(); err != nil {
		return err
	}

	return server.SendAndClose(&protocol.UploadFileResponse{
		Hash: hex.EncodeToString(h.Sum(nil)),
		Size: int64(received),
	})
}

func (s *Storage) GetFile(request *protocol.GetFileRequest, server protocol.Storage_GetFileServer) error {
	file, err := s.fileStorage.Get(server.Context(), request.Id)
	if err != nil {
		return err
	}
	defer file.Close()

	buff := make([]byte, request.ChunkSize)
	for {
		n, err := file.Read(buff)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		err = server.Send(&protocol.GetFileResponse{
			Data: buff[:n],
		})
		if err != nil {
			return err
		}
	}
	return nil
}
