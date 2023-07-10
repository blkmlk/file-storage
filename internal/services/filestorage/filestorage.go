package filestorage

import (
	"context"
	"io"
)

type FileStorage interface {
	Create(ctx context.Context, name string) (io.WriteCloser, error)
	Get(ctx context.Context, name string) (io.ReadCloser, error)
	Exists(ctx context.Context, name string) (bool, error)
}
