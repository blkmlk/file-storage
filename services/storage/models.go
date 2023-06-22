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

type File struct {
	ID        string
	Name      string
	Hash      string
	Status    FileStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewFile(name string) File {
	now := time.Now().UTC()
	return File{
		ID:        uuid.NewString(),
		Name:      name,
		Hash:      "",
		Status:    FileStatusCreated,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
