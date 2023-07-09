package manager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/blkmlk/file-storage/protocol"
)

const (
	ChunkSize = 4096
)

type FilePart struct {
	RemoteID  string
	StorageID string
	Client    protocol.StorageClient
	Size      int64
	Hash      string
}

type loader struct {
	size      int64
	locker    sync.Mutex
	fileParts []FilePart
}

func NewLoader(size int64) *loader {
	return &loader{size: size}
}

func (l *loader) Upload(ctx context.Context, reader io.Reader) error {
	fileParts := l.fileParts

	remainingSize := l.size
	partSize := l.size / int64(len(fileParts))

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

func (l *loader) Download(ctx context.Context, writer io.Writer) error {
	targetSize := l.size
	sent := int64(0)
	fileParts := l.GetFileParts()

	for _, fp := range fileParts {
		n, err := l.getByChunks(ctx, &fp, writer, ChunkSize)
		if err != nil {
			return err
		}
		sent += int64(n)

		if sent > targetSize {
			return fmt.Errorf("sent more than target size")
		}
	}

	if sent != targetSize {
		return fmt.Errorf("sent less than target size")
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
			part.Size = fullSize
		}()

		for data := range dataCh {
			err = stream.Send(&protocol.UploadFileRequest{
				Id:   part.RemoteID,
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

func (l *loader) getByChunks(ctx context.Context, part *FilePart, pipe io.Writer, chunkSize int64) (int, error) {
	var wg sync.WaitGroup
	ch := make(chan []byte, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for data := range ch {
			if _, err := pipe.Write(data); err != nil {
				// todo: handle the err
				return
			}
		}
	}()

	resp, err := part.Client.GetFile(ctx, &protocol.GetFileRequest{Id: part.RemoteID, ChunkSize: ChunkSize})
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = resp.CloseSend()
	}()

	received := 0
	for {
		partData, err := resp.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return 0, err
		}

		ch <- partData.Data
		received += len(partData.Data)

		if int64(received) > part.Size {
			return 0, fmt.Errorf("received more than expected")
		}
	}

	close(ch)
	wg.Wait()

	if int64(received) < part.Size {
		return 0, fmt.Errorf("received less than expected")
	}

	return received, nil
}
