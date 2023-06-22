package storage

import (
	"time"

	"github.com/google/uuid"
)

type FileStatus string

const (
	FileStatusCreated  FileStatus = "created"
	FileStatusUploaded FileStatus = "uploaded"
)

type UploadedFile struct {
	ID        string
	Name      string
	Hash      string
	Status    FileStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUploadedFile(name string) UploadedFile {
	now := time.Now().UTC()
	return UploadedFile{
		ID:        uuid.NewString(),
		Name:      name,
		Hash:      "",
		Status:    FileStatusCreated,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
