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

type FilePart struct {
	ID        string
	Seq       int
	StorageID string
	Hash      string
	Size      int64
}

type Uploader interface {
	Upload(ctx context.Context, reader io.Reader) error
}

type Getter interface {
	Get(ctx context.Context) (io.Reader, error)
}

type GetUploaderInput struct {
	Name        string
	Size        int64
	MinStorages int
	NumStorages int
}

type GetGetterInput struct {
	Name string
}

type Splitter interface {
	GetUploader(ctx context.Context, input GetUploaderInput) (Uploader, error)
	GetGetter(ctx context.Context, input GetGetterInput) (Getter, error)
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
	var cl = newUploader(s.repo, input.Name, input.Size)
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

			client := protocol.NewStorageClient(conn)
			resp, err := client.CheckReadiness(inCtx, &protocol.CheckReadinessRequest{
				Size: size,
			})
			if err != nil || !resp.Ready {
				return
			}
			cl.AddStorage(&Storage{
				StorageID: s.ID,
				Client:    client,
			})
		}(ctx, *s, maxPartSize)
	}
	wg.Wait()

	if cl.LenStorages() < input.MinStorages {
		return nil, ErrNotEnough
	}

	return &cl, nil
}

func (s *splitter) GetGetter(ctx context.Context, input GetGetterInput) (Getter, error) {
	file, err := s.repo.GetFile(ctx, input.Name)
	if err != nil {
		return nil, err
	}

	fileParts, err := s.repo.FindFileParts(ctx, file.ID)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	cl := newUploader(s.repo, file.Name, file.Size)

	for _, fp := range fileParts {
		wg.Add(1)
		go func(filePart *repository.FilePart) {
			defer wg.Done()

			storage, err := s.repo.GetStorage(ctx, filePart.StorageID)
			if err != nil {
				return
			}

			inCtx, cancel := context.WithTimeout(ctx, MaxResponseTime)
			defer cancel()

			conn, err := grpc.DialContext(inCtx, storage.Host)
			if err != nil {
				return
			}
			client := protocol.NewStorageClient(conn)
			resp, err := client.CheckReadiness(inCtx, &protocol.CheckReadinessRequest{
				Size: file.Size,
			})

			if err != nil {
				return
			}

			if !resp.Ready {
				return
			}
			cl.AddStorage(&Storage{
				StorageID: storage.ID,
				Client:    client,
			})
		}(fp)
	}
	wg.Wait()

	if cl.LenStorages() < len(fileParts) {
		return nil, ErrNotEnough
	}

	return &cl, nil
}
