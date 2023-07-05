package splitter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/blkmlk/file-storage/protocol"
	"google.golang.org/grpc"

	"github.com/blkmlk/file-storage/services/repository"
)

const (
	MaxResponseTime = time.Millisecond * 200
)

var (
	ErrNotEnough = errors.New("not enough available storages")
)

type Uploader interface {
	Upload(ctx context.Context, reader io.Reader) error
}

type Loader interface {
}

type GetUploaderInput struct {
	Name        string
	Size        int64
	MinStorages int
	NumStorages int
}

type Splitter interface {
	GetUploader(ctx context.Context, input GetUploaderInput) (Uploader, error)
}

func New(repo repository.Repository) Splitter {
	return &splitter{
		repo: repo,
	}
}

type splitter struct {
	repo repository.Repository
}

func (s *splitter) GetUploader(ctx context.Context, input GetUploaderInput) (Uploader, error) {
	if input.NumStorages < input.MinStorages {
		return nil, fmt.Errorf("num servers is less than min servers")
	}

	storages, err := s.repo.FindStorages(ctx)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	var cl = newCollector(input.Name, input.Size, input.MinStorages)
	maxPartSize := input.Size / int64(input.MinStorages)
	for _, s := range storages {
		wg.Add(1)
		go func(ctx context.Context, s repository.Storage, size int64) {
			inCtx, cancel := context.WithTimeout(ctx, MaxResponseTime)

			defer cancel()
			defer wg.Done()

			conn, err := grpc.DialContext(inCtx, s.Host)
			if err != nil {
				return
			}

			storage := protocol.NewStorageClient(conn)
			resp, err := storage.CheckReadiness(inCtx, &protocol.CheckReadinessRequest{
				Size: size,
			})
			if err != nil || !resp.Ready {
				return
			}
			cl.AddStorage(storage)
		}(ctx, *s, maxPartSize)
	}
	wg.Wait()

	if cl.LenStorages() < input.MinStorages {
		return nil, ErrNotEnough
	}

	return &cl, nil
}
