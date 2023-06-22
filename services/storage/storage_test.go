package storage_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/blkmlk/file-storage/deps"
	"github.com/blkmlk/file-storage/migrations"
	"github.com/blkmlk/file-storage/services/storage"
	"github.com/stretchr/testify/suite"
	"go.uber.org/dig"
)

func TestAll(t *testing.T) {
	suite.Run(t, new(testSuite))
}

type testSuite struct {
	suite.Suite
	ctn     *dig.Container
	storage storage.Storage
}

func (t *testSuite) SetupTest() {
	t.ctn = dig.New()
	t.Require().NoError(t.ctn.Provide(deps.NewLocalDB))
	t.Require().NoError(t.ctn.Provide(storage.New))

	m, err := migrations.NewLocal()
	t.Require().NoError(err)

	if m.Up() != nil {
		t.Require().NoError(m.Drop())
		m, _ = migrations.NewLocal()
		t.Require().NoError(m.Up())
	}

	err = t.ctn.Invoke(func(s storage.Storage) {
		t.storage = s
	})
	t.Require().NoError(err)
}

func (t *testSuite) TestCreateUpdateAndGetFile() {
	ctx := context.Background()
	file := storage.NewFile("test")

	err := t.storage.CreateFile(ctx, &file)
	t.Require().NoError(err)

	err = t.storage.CreateFile(ctx, &file)
	t.Require().ErrorIs(err, storage.ErrAlreadyExists)

	err = t.storage.UpdateFileStatus(ctx, file.ID, "hash", storage.FileStatusUploaded)
	t.Require().NoError(err)

	err = t.storage.UpdateFileStatus(ctx, uuid.NewString(), "hash", storage.FileStatusUploaded)
	t.Require().ErrorIs(err, storage.ErrNotFound)

	foundFile, err := t.storage.GetFile(ctx, file.Name)
	t.Require().NoError(err)
	t.Require().Equal(storage.FileStatusUploaded, foundFile.Status)

	foundFile, err = t.storage.GetFile(ctx, "unknown")
	t.Require().ErrorIs(err, storage.ErrNotFound)
	t.Require().Nil(foundFile)
}

func (t *testSuite) TestCreateFileParts() {
	ctx := context.Background()
	file := storage.NewFile("test")

	err := t.storage.CreateFile(ctx, &file)
	t.Require().NoError(err)

	fileStorage := storage.NewFileStorage(uuid.NewString())
	err = t.storage.CreateFileStorage(ctx, &fileStorage)
	t.Require().NoError(err)

	for i := 0; i < 10; i++ {
		filePart := storage.NewFilePart(file.ID, i, fileStorage.ID, uuid.NewString())
		err = t.storage.CreateFilePart(ctx, &filePart)
		t.Require().NoError(err)
	}

	foundFileParts, err := t.storage.FindFileParts(ctx, file.ID)
	t.Require().NoError(err)
	t.Require().Len(foundFileParts, 10)
}
