package storage

import (
	"context"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var (
	ErrAlreadyExists = errors.New("already exists")
)

type Storage interface {
	CreateUploadedFile(ctx context.Context, file *UploadedFile) error
}

type storage struct {
	db *gorm.DB
}

func New(db *gorm.DB) Storage {
	s := storage{
		db: db,
	}

	return s
}

func (s storage) CreateUploadedFile(ctx context.Context, file *UploadedFile) error {
	tx := s.db.WithContext(ctx).Create(file)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrDuplicatedKey) {
			return ErrAlreadyExists
		}
		return tx.Error
	}
	return nil
}
