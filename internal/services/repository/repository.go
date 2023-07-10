package repository

import (
	"context"
	"time"

	"gorm.io/gorm/clause"

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

type Repository interface {
	CreateFile(ctx context.Context, file *File) error
	UpdateFileStatus(ctx context.Context, id, name string, size int64, status FileStatus) error
	GetFile(ctx context.Context, id string) (*File, error)
	GetFileByName(ctx context.Context, name string) (*File, error)

	CreateOrUpdateStorage(ctx context.Context, storage *Storage) error
	GetStorage(ctx context.Context, id string) (*Storage, error)
	FindStorages(ctx context.Context) ([]*Storage, error)

	CreateFilePart(ctx context.Context, filePart *FilePart) error
	CreateFileParts(ctx context.Context, fileParts []FilePart) error
	FindOrderedFileParts(ctx context.Context, fileID string) ([]*FilePart, error)
}

type storage struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &storage{
		db: db,
	}
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

func (s storage) UpdateFileStatus(ctx context.Context, id, name string, size int64, status FileStatus) error {
	tx := s.db.WithContext(ctx).Table("files").Where("id = ?", id).
		Updates(map[string]any{
			"name":   name,
			"size":   size,
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

func (s storage) GetFile(ctx context.Context, id string) (*File, error) {
	var file File
	tx := s.db.WithContext(ctx).Table("files").Where("id = ?", id).Find(&file)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}

	return &file, nil
}

func (s storage) GetFileByName(ctx context.Context, name string) (*File, error) {
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

func (s storage) CreateOrUpdateStorage(ctx context.Context, fileStorage *Storage) error {
	return s.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: []clause.Assignment{
			{
				Column: clause.Column{Name: "host"},
				Value:  fileStorage.Host,
			},
			{
				Column: clause.Column{Name: "updated_at"},
				Value:  time.Now(),
			},
		},
	}).WithContext(ctx).Create(fileStorage).Error
}

func (s storage) GetStorage(ctx context.Context, id string) (*Storage, error) {
	//TODO implement me
	panic("implement me")
}

func (s storage) FindStorages(ctx context.Context) ([]*Storage, error) {
	var result []*Storage
	if err := s.db.WithContext(ctx).Table("storages").Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
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

func (s storage) CreateFileParts(ctx context.Context, fileParts []FilePart) error {
	tx := s.db.WithContext(ctx).CreateInBatches(fileParts, len(fileParts))
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (s storage) FindOrderedFileParts(ctx context.Context, fileID string) ([]*FilePart, error) {
	var fileParts []*FilePart
	tx := s.db.WithContext(ctx).Table("file_parts").
		Where("file_id = ?", fileID).Order("seq ASC").Find(&fileParts)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return fileParts, nil
}
