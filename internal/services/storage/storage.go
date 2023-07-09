package storage

import (
	"github.com/blkmlk/file-storage/internal/services/repository"
)

type Storage interface {
	SaveFile(file []byte) error
}

type storage struct {
	repo repository.Repository
}

func New(repo repository.Repository) (Storage, error) {
	s := storage{
		repo: repo,
	}

	return s, nil
}

func (s storage) SaveFile(file []byte) error {
	//TODO implement me
	panic("implement me")
}
