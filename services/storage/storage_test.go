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

func (t *testSuite) TestCreateFile() {
	ctx := context.Background()
	file := storage.NewUploadedFile("test")

	err := t.storage.CreateUploadedFile(ctx, &file)
	t.Require().NoError(err)

	err = t.storage.CreateUploadedFile(ctx, &file)
	t.Require().ErrorIs(err, storage.ErrAlreadyExists)

	err = t.storage.UpdateUploadedFileStatus(ctx, file.ID, "hash", storage.FileStatusUploaded)
	t.Require().NoError(err)

	err = t.storage.UpdateUploadedFileStatus(ctx, uuid.NewString(), "hash", storage.FileStatusUploaded)
	t.Require().ErrorIs(err, storage.ErrNotFound)
}
