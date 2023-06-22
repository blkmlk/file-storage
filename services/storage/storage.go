package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var (
	ErrAlreadyExists = errors.New("already exists")
	ErrNotFound      = errors.New("not found")
)

type Storage interface {
	CreateUploadedFile(ctx context.Context, file *UploadedFile) error
	UpdateUploadedFileStatus(ctx context.Context, fileID string, hash string, status FileStatus) error
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
		if e, ok := tx.Error.(*pgconn.PgError); ok && e.Code == "23505" {
			return ErrAlreadyExists
		}
		return tx.Error
	}
	return nil
}

func (s storage) UpdateUploadedFileStatus(ctx context.Context, fileID string, hash string, status FileStatus) error {
	tx := s.db.WithContext(ctx).Table("uploaded_files").Where("id = ?", fileID).
		Updates(map[string]any{
			"hash":   hash,
			"status": status,
		})

	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
