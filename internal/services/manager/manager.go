package manager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/blkmlk/file-storage/internal/helpers"

	"go.uber.org/zap"

	"github.com/blkmlk/file-storage/env"
	_ "github.com/hashicorp/go-multierror"

	"github.com/blkmlk/file-storage/internal/services/cache"
	"github.com/blkmlk/file-storage/internal/services/repository"
	"github.com/blkmlk/file-storage/protocol"
)

const (
	MaxResponseTime = time.Millisecond * 200
)

var (
	ErrBusy     = errors.New("file is busy")
	ErrExists   = errors.New("file exists")
	ErrNotFound = errors.New("not found")
)

type FileInfo struct {
	Name        string
	ContentType string
	Size        int64
}

type Manager interface {
	Prepare(ctx context.Context) (string, error)
	Store(ctx context.Context, id string, info FileInfo, reader io.Reader) error
	Load(ctx context.Context, name string) (io.Reader, error)
}

func New(
	repo repository.Repository,
	cache cache.Cache,
	clientFactory ClientFactory,
) (Manager, error) {
	value, err := env.Get(env.MinStorages)
	if err != nil {
		return nil, err
	}

	minStorages, err := strconv.Atoi(value)
	if err != nil {
		return nil, fmt.Errorf("%s is not integer", env.MinStorages)
	}

	return &manager{
		cache:         cache,
		repo:          repo,
		clientFactory: clientFactory,
		minStorages:   minStorages,
	}, nil
}

type manager struct {
	log           *zap.SugaredLogger
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
		return ErrBusy
	}
	defer m.cache.Unlock(keys)

	if _, err := m.repo.GetFileByName(ctx, info.Name); err != nil && !errors.Is(err, repository.ErrNotFound) {
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
		return ErrExists
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

	if err = m.repo.UpdateFileInfo(ctx, file.ID, repository.UpdateFileInfoInput{
		Name:        info.Name,
		ContentType: info.ContentType,
		Size:        info.Size,
		Status:      repository.FileStatusUploaded,
	}); err != nil {
		return err
	}

	return nil
}

func (m *manager) Load(ctx context.Context, name string) (io.Reader, error) {
	file, err := m.repo.GetFileByName(ctx, name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	ldr, err := m.prepareLoaderForDownload(ctx, file)
	if err != nil {
		return nil, err
	}

	return ldr.Download(ctx)
}

func (m *manager) prepareLoaderForUpload(ctx context.Context, info FileInfo) (*loader, error) {
	storages, err := m.repo.FindStorages(ctx)
	if err != nil {
		return nil, err
	}

	if len(storages) < m.minStorages {
		return nil, fmt.Errorf("not enough storages")
	}

	ldr := NewLoader(m.log, info.Size)

	var wg sync.WaitGroup
	maxPartSize := info.Size / int64(m.minStorages)
	errs := make(chan error, len(storages))
	for _, s := range storages {
		wg.Add(1)
		go func(ctx context.Context, s repository.Storage, size int64) {
			defer wg.Done()

			client, err := m.clientFactory.NewStorageClient(ctx, s.Host)
			if err != nil {
				errs <- fmt.Errorf("failed to connect to storage (%s): %v", s.ID, err)
				return
			}
			reqCtx, cancel := context.WithTimeout(ctx, MaxResponseTime)
			defer cancel()

			resp, err := client.CheckReadiness(reqCtx, &protocol.CheckReadinessRequest{
				Size: size,
			})
			if err != nil {
				errs <- fmt.Errorf("failed to check storage rediness (%s): %v", s.ID, err)
				return
			}
			if !resp.Ready {
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

	close(errs)
	if err = helpers.ReadErrors(errs); err != nil {
		return nil, err
	}

	if ldr.LenFileParts() < m.minStorages {
		return nil, fmt.Errorf("not enough file parts")
	}

	return ldr, nil
}

func (m *manager) prepareLoaderForDownload(ctx context.Context, file *repository.File) (*loader, error) {
	fileParts, err := m.repo.FindFileParts(ctx, file.ID)
	if err != nil {
		return nil, err
	}

	ldr := NewLoader(m.log, file.Size)

	var wg sync.WaitGroup
	errs := make(chan error, len(fileParts))
	for i := 0; i < len(fileParts); i++ {
		wg.Add(1)
		go func(ctx context.Context, fp repository.FilePart) {
			defer wg.Done()

			storage, err := m.repo.GetStorage(ctx, fp.StorageID)
			if err != nil {
				errs <- fmt.Errorf("failed to get storage from DB: %v", err)
				return
			}

			client, err := m.clientFactory.NewStorageClient(ctx, storage.Host)
			if err != nil {
				errs <- fmt.Errorf("failed to connect to storage (%s): %v", storage.ID, err)
				return
			}
			reqCtx, cancel := context.WithTimeout(ctx, MaxResponseTime)
			defer cancel()

			resp, err := client.CheckFilePartExistence(reqCtx, &protocol.CheckFilePartExistenceRequest{
				Id: fp.RemoteID,
			})
			if err != nil {
				errs <- fmt.Errorf("failed to check file part (%s): %v", storage.ID, err)
				return
			}

			if !resp.Exists {
				errs <- fmt.Errorf("file part doens't exist: %s", fp.ID)
				return
			}

			ldr.AddFilePart(&FilePart{
				Seq:       fp.Seq,
				StorageID: storage.ID,
				Client:    client,
				RemoteID:  fp.RemoteID,
				Size:      fp.Size,
				Hash:      fp.Hash,
			})
		}(ctx, *fileParts[i])
	}
	wg.Wait()

	close(errs)
	if err = helpers.ReadErrors(errs); err != nil {
		return nil, err
	}

	if ldr.LenFileParts() != len(fileParts) {
		return nil, fmt.Errorf("not enough parts")
	}

	ldr.SortFileParts()

	return ldr, nil
}
