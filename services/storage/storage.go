package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const (
	ConstraintErrorCode = "23505"
)

var (
	ErrAlreadyExists = errors.New("already exists")
	ErrNotFound      = errors.New("not found")
)

type Storage interface {
	CreateFile(ctx context.Context, file *File) error
	UpdateFileStatus(ctx context.Context, fileID string, hash string, status FileStatus) error
	GetFile(ctx context.Context, name string) (*File, error)

	CreateFileStorage(ctx context.Context, fileStorage *FileStorage) error

	CreateFilePart(ctx context.Context, filePart *FilePart) error
	FindFileParts(ctx context.Context, fileID string) ([]*FilePart, error)
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

func (s storage) CreateFile(ctx context.Context, file *File) error {
	tx := s.db.WithContext(ctx).Create(file)
	if tx.Error != nil {
		if e, ok := tx.Error.(*pgconn.PgError); ok && e.Code == ConstraintErrorCode {
			return ErrAlreadyExists
		}
		return tx.Error
	}
	return nil
}

func (s storage) UpdateFileStatus(ctx context.Context, fileID string, hash string, status FileStatus) error {
	tx := s.db.WithContext(ctx).Table("files").Where("id = ?", fileID).
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

func (s storage) GetFile(ctx context.Context, name string) (*File, error) {
	var file File
	tx := s.db.WithContext(ctx).Table("files").Where("name = ?", name).Find(&file)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}

	return &file, nil
}

func (s storage) CreateFileStorage(ctx context.Context, fileStorage *FileStorage) error {
	tx := s.db.WithContext(ctx).Table("storages").Create(fileStorage)
	if tx.Error != nil {
		if e, ok := tx.Error.(*pgconn.PgError); ok && e.Code == ConstraintErrorCode {
			return ErrAlreadyExists
		}
		return tx.Error
	}
	return nil
}

func (s storage) CreateFilePart(ctx context.Context, filePart *FilePart) error {
	tx := s.db.WithContext(ctx).Create(filePart)
	if tx.Error != nil {
		if e, ok := tx.Error.(*pgconn.PgError); ok && e.Code == ConstraintErrorCode {
			return ErrAlreadyExists
		}
		return tx.Error
	}
	return nil
}

func (s storage) FindFileParts(ctx context.Context, fileID string) ([]*FilePart, error) {
	var fileParts []*FilePart
	tx := s.db.WithContext(ctx).Table("file_parts").
		Where("file_id = ?", fileID).Order("seq ASC").Find(&fileParts)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return fileParts, nil
}
