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
	Storages []protocol.StorageClient
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

		checkResp, err := storages[i].CheckReadiness(ctx, &protocol.CheckReadinessRequest{
			Size: size,
		})
		if err != nil {
			return nil, err
		}

		if !checkResp.Ready {
			return nil, fmt.Errorf("storage is not ready")
		}

		if err = sendByChunks(ctx, storages[i], reader, checkResp.Id, size, ChunkSize); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func sendByChunks(ctx context.Context, storage protocol.StorageClient, pipe io.Reader, id string, fullSize, chunkSize int64) error {
	inCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	buff := make([]byte, chunkSize)
	remainingSize := fullSize

	dataCh := make(chan []byte, 1)
	errCh := make(chan error, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()
		stream, err := storage.UploadFile(inCtx)
		if err != nil {
			errCh <- err
			return
		}

		defer func() {
			if err = stream.CloseSend(); err != nil {
				errCh <- err
				return
			}
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
		return <-errCh
	}

	return gErr
}

func newCollector(name string, size int64, minStorages int) collector {
	return collector{
		Name:        name,
		Size:        size,
		MinStorages: minStorages,
	}
}

func (c *collector) AddStorage(s protocol.StorageClient) {
	c.locker.Lock()
	defer c.locker.Unlock()
	c.Storages = append(c.Storages, s)
}

func (c *collector) GetStorages() []protocol.StorageClient {
	c.locker.Lock()
	defer c.locker.Unlock()

	copied := make([]protocol.StorageClient, len(c.Storages))
	copy(copied, c.Storages)
	return copied
}

func (c *collector) LenStorages() int {
	c.locker.Lock()
	defer c.locker.Unlock()
	return len(c.Storages)
}
