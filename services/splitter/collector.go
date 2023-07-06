package splitter

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/blkmlk/file-storage/protocol"
)

const (
	ChunkSize = 4096
)

type collector struct {
	Name        string
	Size        int64
	MinStorages int

	locker   sync.Mutex
	Storages []Storage
}

type Storage struct {
	StorageID string
	Client    protocol.StorageClient
}

func (c *collector) Upload(ctx context.Context, reader io.Reader) ([]*FilePart, error) {
	storages := c.GetStorages()
	var result []*FilePart

	remainingSize := c.Size
	partSize := c.Size / int64(len(storages))

	for i := 0; i < len(storages); i++ {
		last := i == len(storages)-1

		size := partSize
		if last {
			size = remainingSize
		}

		checkResp, err := storages[i].Client.CheckReadiness(ctx, &protocol.CheckReadinessRequest{
			Size: size,
		})
		if err != nil {
			return nil, err
		}

		if !checkResp.Ready {
			return nil, fmt.Errorf("storage is not ready")
		}

		filePart, err := sendByChunks(ctx, &storages[i], reader, checkResp.Id, size, ChunkSize)
		if err != nil {
			return nil, err
		}
		filePart.Seq = i
		filePart.Size = size
		filePart.StorageID = storages[i].StorageID

		result = append(result, filePart)
	}

	return result, nil
}

func sendByChunks(ctx context.Context, storage *Storage, pipe io.Reader, id string, fullSize, chunkSize int64) (*FilePart, error) {
	inCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	buff := make([]byte, chunkSize)
	remainingSize := fullSize

	dataCh := make(chan []byte, 1)
	errCh := make(chan error, 2)
	var filePart FilePart

	wg.Add(1)
	go func() {
		defer wg.Done()
		stream, err := storage.Client.UploadFile(inCtx)
		if err != nil {
			errCh <- err
			return
		}

		defer func() {
			resp, err := stream.CloseAndRecv()
			if err != nil {
				errCh <- err
				return
			}
			filePart.ID = resp.Id
			filePart.Hash = resp.Hash
		}()

		for data := range dataCh {
			err = stream.Send(&protocol.UploadFileRequest{
				Id:   id,
				Data: data,
			})
			if err != nil {
				errCh <- err
				return
			}
		}
	}()

	var gErr error

	for remainingSize > 0 {
		chunk := chunkSize
		if remainingSize < chunk {
			chunk = remainingSize
		}

		data := buff[:chunk]
		n, err := pipe.Read(data)
		if err != nil {
			gErr = err
			break
		}

		if int64(n) < chunk {
			gErr = fmt.Errorf("read less than expected")
			break
		}

		select {
		case <-ctx.Done():
			gErr = ctx.Err()
			break
		case gErr = <-errCh:
			break
		case dataCh <- data[:n]:
		}

		remainingSize -= chunk
	}

	if gErr != nil {
		cancel()
	}

	close(dataCh)
	wg.Wait()
	close(errCh)

	if len(errCh) > 0 {
		gErr = <-errCh
	}

	if gErr != nil {
		return nil, gErr
	}

	return &filePart, nil
}

func newCollector(name string, size int64, minStorages int) collector {
	return collector{
		Name:        name,
		Size:        size,
		MinStorages: minStorages,
	}
}

func (c *collector) AddStorage(s *Storage) {
	c.locker.Lock()
	defer c.locker.Unlock()
	c.Storages = append(c.Storages, *s)
}

func (c *collector) GetStorages() []Storage {
	c.locker.Lock()
	defer c.locker.Unlock()

	copied := make([]Storage, len(c.Storages))
	copy(copied, c.Storages)
	return copied
}

func (c *collector) LenStorages() int {
	c.locker.Lock()
	defer c.locker.Unlock()
	return len(c.Storages)
}
