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
	Name      *string
	Hash      string
	Size      int64
	Status    FileStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewFile() File {
	now := time.Now().UTC()
	return File{
		ID:        uuid.NewString(),
		Name:      nil,
		Hash:      "",
		Status:    FileStatusCreated,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type FilePart struct {
	ID        string
	FileID    string
	RemoteID  string
	Seq       int
	Hash      string
	StorageID string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewFilePart(fileID, remoteID string, seq int, storageID, hash string) FilePart {
	now := time.Now()
	return FilePart{
		ID:        uuid.NewString(),
		FileID:    fileID,
		RemoteID:  remoteID,
		Seq:       seq,
		Hash:      hash,
		StorageID: storageID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type Storage struct {
	ID        string
	Host      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewStorage(id, host string) Storage {
	return Storage{
		ID:        id,
		Host:      host,
		CreatedAt: time.Now(),
	}
}
