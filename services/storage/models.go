package storage

import (
	"github.com/google/uuid"
	"time"
)

type UploadedFile struct {
	ID        string
	Name      string
	Hash      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUploadedFile(name string) UploadedFile {
	now := time.Now().UTC()
	return UploadedFile{
		ID:        uuid.NewString(),
		Name:      name,
		Hash:      "",
		CreatedAt: now,
		UpdatedAt: now,
	}
}
