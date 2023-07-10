package manager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/blkmlk/file-storage/internal/services/cache"
	"github.com/blkmlk/file-storage/internal/services/repository"
	"github.com/blkmlk/file-storage/protocol"
)

const (
	MaxResponseTime = time.Millisecond * 200
)

var (
	ErrExists   = errors.New("file exists")
	ErrNotFound = errors.New("not found")
)

type FileInfo struct {
	Name string
	Size int64
}

type Manager interface {
	Prepare(ctx context.Context) (string, error)
	Store(ctx context.Context, id string, info FileInfo, reader io.Reader) error
	Load(ctx context.Context, name string, writer io.Writer) error
}

func New(
	repo repository.Repository,
	cache cache.Cache,
	clientFactory ClientFactory,
) Manager {
	return &manager{
		cache:         cache,
		repo:          repo,
		clientFactory: clientFactory,
	}
}

type manager struct {
	repo          repository.Repository
	cache         cache.Cache
	clientFactory ClientFactory
	minStorages   int
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
	keys := []string{fileID, info.Name}
	if err := m.cache.Lock(keys); err != nil {
		return fmt.Errorf("file is already being stored")
	}
	defer m.cache.Unlock(keys)

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
		part := repository.NewFilePart(file.ID, fp.RemoteID, seq, fp.Size, fp.StorageID, fp.Hash)
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

func (m *manager) Load(ctx context.Context, name string, writer io.Writer) error {
	file, err := m.repo.GetFileByName(ctx, name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}

	ldr, err := m.prepareLoaderForDownload(ctx, file)
	if err != nil {
		return err
	}

	return ldr.Download(ctx, writer)
}

func (m *manager) prepareLoaderForUpload(ctx context.Context, info FileInfo) (*loader, error) {
	storages, err := m.repo.FindStorages(ctx)
	if err != nil {
		return nil, err
	}

	if len(storages) < m.minStorages {
		return nil, fmt.Errorf("not enough storages")
	}

	connCtx, connCancel := context.WithCancel(ctx)
	defer connCancel()

	ldr := NewLoader(info.Size)

	var wg sync.WaitGroup
	maxPartSize := info.Size / int64(m.minStorages)
	for _, s := range storages {
		wg.Add(1)
		go func(ctx context.Context, s repository.Storage, size int64) {
			defer wg.Done()

			client, err := m.clientFactory.NewStorageClient(connCtx, s.Host)
			if err != nil {
				return
			}
			reqCtx, cancel := context.WithTimeout(connCtx, MaxResponseTime)
			defer cancel()

			resp, err := client.CheckReadiness(reqCtx, &protocol.CheckReadinessRequest{
				Size: size,
			})
			if err != nil || !resp.Ready {
				return
			}
			ldr.AddFilePart(&FilePart{
				RemoteID:  resp.Id,
				StorageID: s.ID,
				Client:    client,
			})
		}(ctx, *s, maxPartSize)
	}
	wg.Wait()

	if connCtx.Err() != nil {
		return nil, connCtx.Err()
	}

	if ldr.LenFileParts() < m.minStorages {
		return nil, fmt.Errorf("not enough storages")
	}

	return ldr, nil
}

func (m *manager) prepareLoaderForDownload(ctx context.Context, file *repository.File) (*loader, error) {
	fileParts, err := m.repo.FindOrderedFileParts(ctx, file.ID)
	if err != nil {
		return nil, err
	}

	connCtx, connCancel := context.WithCancel(ctx)
	defer connCancel()

	ldr := NewLoader(file.Size)

	var wg sync.WaitGroup
	for i := 0; i < len(fileParts); i++ {
		wg.Add(1)
		go func(ctx context.Context, fp repository.FilePart) {
			defer wg.Done()

			storage, err := m.repo.GetStorage(ctx, fp.StorageID)
			if err != nil {
				return
			}

			client, err := m.clientFactory.NewStorageClient(connCtx, storage.Host)
			if err != nil {
				return
			}
			reqCtx, cancel := context.WithTimeout(connCtx, MaxResponseTime)
			defer cancel()

			resp, err := client.CheckFilePartExistence(reqCtx, &protocol.CheckFilePartExistenceRequest{
				Id: fp.RemoteID,
			})
			if err != nil {
				return
			}

			if !resp.Exists {
				return
			}

			ldr.AddFilePart(&FilePart{
				StorageID: storage.ID,
				Client:    client,
				RemoteID:  fp.RemoteID,
				Size:      fp.Size,
				Hash:      fp.Hash,
			})
		}(ctx, *fileParts[i])
	}
	wg.Wait()

	if connCtx.Err() != nil {
		return nil, connCtx.Err()
	}

	if ldr.LenFileParts() != len(fileParts) {
		return nil, fmt.Errorf("not enough parts")
	}

	return ldr, nil
}
