package filestorage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/blkmlk/file-storage/env"
)

var (
	ErrAlreadyExists = errors.New("already exists")
	ErrNotFound      = errors.New("not found")
)

type fsFileStorage struct {
	rootPath string
}

func NewFS() (FileStorage, error) {
	rootPath, err := env.Get(env.FSRootPath)
	if err != nil {
		return nil, err
	}

	return &fsFileStorage{
		rootPath: rootPath,
	}, nil
}

func (f *fsFileStorage) Create(ctx context.Context, name string) (io.WriteCloser, error) {
	filePath := f.getFilePath(name)

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0700)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil, ErrAlreadyExists
		}
		return nil, err
	}
	return file, nil
}

func (f *fsFileStorage) Get(ctx context.Context, name string) (io.ReadCloser, error) {
	filePath := f.getFilePath(name)

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0700)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return file, nil
}

func (f *fsFileStorage) Exists(ctx context.Context, name string) (bool, error) {
	filePath := f.getFilePath(name)

	info, err := os.Stat(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return !info.IsDir(), nil
}

func (f *fsFileStorage) getFilePath(name string) string {
	return fmt.Sprintf("%s/%s", f.rootPath, name)
}
