package manager

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
	ErrExists = errors.New("file exists")
)

type FileInfo struct {
	Name string
	Size int64
}

type Manager interface {
	Prepare(ctx context.Context) (string, error)
	Store(ctx context.Context, id string, info FileInfo, reader io.Reader) error
	Load(ctx context.Context, name string) (io.Reader, error)
}

func New(repo repository.Repository) Manager {
	return &manager{
		repo: repo,
	}
}

type manager struct {
	repo        repository.Repository
	minStorages int
}

func (m *manager) Prepare(ctx context.Context) (string, error) {
	newFile := repository.NewFile()
	if err := m.repo.CreateFile(ctx, &newFile); err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return "", ErrExists
		}
		return "", err
	}
	return newFile.ID, nil
}

func (m *manager) Store(ctx context.Context, fileID string, info FileInfo, reader io.Reader) error {
	// todo: add cache name and id locks
	if _, err := m.repo.GetFileByName(ctx, info.Name); err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return ErrExists
		}
		return err
	}

	file, err := m.repo.GetFile(ctx, fileID)
	if err != nil {
		return err
	}

	if file.Status != repository.FileStatusCreated {
		return fmt.Errorf("wrong status")
	}

	ldr, err := m.prepareLoaderForUpload(ctx, info)
	if err != nil {
		return err
	}

	if err = ldr.Upload(ctx, reader); err != nil {
		return err
	}

	var dbFileParts = make([]repository.FilePart, 0, ldr.LenFileParts())
	for seq, fp := range ldr.GetFileParts() {
		part := repository.NewFilePart(file.ID, fp.RemoteID, seq, fp.StorageID, fp.Hash)
		dbFileParts = append(dbFileParts, part)
	}

	if err = m.repo.CreateFileParts(ctx, dbFileParts); err != nil {
		return err
	}

	if err = m.repo.UpdateFileStatus(ctx, file.ID, info.Name, info.Size, repository.FileStatusUploaded); err != nil {
		return err
	}

	return nil
}

func (m *manager) Load(ctx context.Context, name string) (io.Reader, error) {
	//TODO implement me
	panic("implement me")
}

func (m *manager) prepareLoaderForUpload(ctx context.Context, info FileInfo) (*loader, error) {
	storages, err := m.repo.FindStorages(ctx)
	if err != nil {
		return nil, err
	}

	if len(storages) < m.minStorages {
		return nil, fmt.Errorf("not enough storages")
	}

	ldr := NewLoader(info)

	var wg sync.WaitGroup
	maxPartSize := info.Size / int64(m.minStorages)
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
			ldr.AddFilePart(&FilePart{
				ID:        resp.Id,
				StorageID: s.ID,
				Client:    client,
			})
		}(ctx, *s, maxPartSize)
	}
	wg.Wait()

	if ldr.LenFileParts() < m.minStorages {
		return nil, fmt.Errorf("not enough storages")
	}

	return ldr, nil
}
