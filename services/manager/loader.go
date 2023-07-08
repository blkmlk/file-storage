package manager

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

type FilePart struct {
	ID        string
	StorageID string
	Client    protocol.StorageClient

	RemoteID string
	Hash     string
}

type loader struct {
	info      FileInfo
	locker    sync.Mutex
	fileParts []FilePart
}

func NewLoader(info FileInfo) *loader {
	return &loader{info: info}
}

func (l *loader) Upload(ctx context.Context, reader io.Reader) error {
	fileParts := l.fileParts

	remainingSize := l.info.Size
	partSize := l.info.Size / int64(len(fileParts))

	for i := 0; i < len(fileParts); i++ {
		last := i == len(fileParts)-1

		size := partSize
		if last {
			size = remainingSize
		}

		err := l.sendByChunks(ctx, &fileParts[i], reader, size, ChunkSize)
		if err != nil {
			return err
		}

		remainingSize -= size
	}

	return nil
}

func (l *loader) AddFilePart(fp *FilePart) {
	l.locker.Lock()
	defer l.locker.Unlock()
	l.fileParts = append(l.fileParts, *fp)
}

func (l *loader) GetFileParts() []FilePart {
	l.locker.Lock()
	defer l.locker.Unlock()

	copied := make([]FilePart, len(l.fileParts))
	copy(copied, l.fileParts)
	return copied
}

func (l *loader) LenFileParts() int {
	l.locker.Lock()
	defer l.locker.Unlock()
	return len(l.fileParts)
}

func (l *loader) sendByChunks(ctx context.Context, part *FilePart, pipe io.Reader, fullSize, chunkSize int64) error {
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
		stream, err := part.Client.UploadFile(inCtx)
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
			part.RemoteID = resp.Id
			part.Hash = resp.Hash
		}()

		for data := range dataCh {
			err = stream.Send(&protocol.UploadFileRequest{
				Id:   part.ID,
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

	return gErr
}
