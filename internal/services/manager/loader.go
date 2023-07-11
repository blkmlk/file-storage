package manager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"sync"

	"github.com/blkmlk/file-storage/protocol"
)

const (
	ChunkSize = 4096
)

type FilePart struct {
	Seq       int
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
	l.locker.Lock()
	defer l.locker.Unlock()

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

func (l *loader) Download(ctx context.Context) (io.Reader, error) {
	l.locker.Lock()
	defer l.locker.Unlock()

	fileParts := l.fileParts

	readers := make([]io.Reader, 0, len(fileParts))
	for _, fp := range fileParts {
		r, err := l.getByChunks(ctx, &fp, ChunkSize)
		if err != nil {
			return nil, err
		}

		readers = append(readers, r)
	}
	reader := io.MultiReader(readers...)

	return reader, nil
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

func (l *loader) SortFileParts() {
	l.locker.Lock()
	defer l.locker.Unlock()
	sort.Slice(l.fileParts, func(i, j int) bool {
		return l.fileParts[i].Seq < l.fileParts[j].Seq
	})
}

func (l *loader) LenFileParts() int {
	l.locker.Lock()
	defer l.locker.Unlock()
	return len(l.fileParts)
}

func (l *loader) sendByChunks(ctx context.Context, part *FilePart, reader io.Reader, fullSize, chunkSize int64) error {
	inCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup

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
	remainingSize := fullSize
	buff := make([]byte, chunkSize)

	for remainingSize > 0 {
		chunk := chunkSize
		if remainingSize < chunk {
			chunk = remainingSize
		}

		data := buff[:chunk]
		n, err := reader.Read(data)
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

func (l *loader) getByChunks(ctx context.Context, part *FilePart, chunkSize int64) (io.Reader, error) {
	r, w := io.Pipe()

	resp, err := part.Client.GetFile(ctx, &protocol.GetFileRequest{Id: part.RemoteID, ChunkSize: chunkSize})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.CloseSend()
	}()

	go func() {
		defer w.Close()
		for {
			partData, err := resp.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				//return nil, err
				return
			}

			if _, err = w.Write(partData.Data); err != nil {
				//return nil, err
				return
			}
		}
	}()

	return r, nil
}
