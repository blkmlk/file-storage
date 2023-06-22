package repository

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

type FilePart struct {
	ID        string
	FileID    string
	Seq       int
	Hash      string
	StorageID string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewFilePart(fileID string, seq int, storageID, hash string) FilePart {
	now := time.Now()
	return FilePart{
		ID:        uuid.NewString(),
		FileID:    fileID,
		Seq:       seq,
		Hash:      hash,
		StorageID: storageID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type FileStorage struct {
	ID        string
	CreatedAt time.Time
}

func NewFileStorage(id string) FileStorage {
	return FileStorage{
		ID:        id,
		CreatedAt: time.Now(),
	}
}
