package repository_test

import (
	"context"
	"testing"

	repository2 "github.com/blkmlk/file-storage/internal/services/repository"

	"github.com/google/uuid"

	"github.com/blkmlk/file-storage/deps"
	"github.com/blkmlk/file-storage/migrations"
	"github.com/stretchr/testify/suite"
	"go.uber.org/dig"
)

func TestAll(t *testing.T) {
	suite.Run(t, new(testSuite))
}

type testSuite struct {
	suite.Suite
	ctn        *dig.Container
	repository repository2.Repository
}

func (t *testSuite) SetupTest() {
	t.ctn = dig.New()
	t.Require().NoError(t.ctn.Provide(deps.NewLocalDB))
	t.Require().NoError(t.ctn.Provide(repository2.New))

	m, err := migrations.NewLocal()
	t.Require().NoError(err)

	if m.Up() != nil {
		t.Require().NoError(m.Drop())
		m, _ = migrations.NewLocal()
		t.Require().NoError(m.Up())
	}

	err = t.ctn.Invoke(func(repo repository2.Repository) {
		t.repository = repo
	})
	t.Require().NoError(err)
}

func (t *testSuite) TestCreateUpdateAndGetFile() {
	ctx := context.Background()
	file := repository2.NewFile()

	err := t.repository.CreateFile(ctx, &file)
	t.Require().NoError(err)

	err = t.repository.CreateFile(ctx, &file)
	t.Require().ErrorIs(err, repository2.ErrAlreadyExists)

	err = t.repository.UpdateFileInfo(ctx, file.ID, repository2.UpdateFileInfoInput{
		Name:        "name-1",
		ContentType: "application/zip",
		Size:        100,
		Status:      repository2.FileStatusUploaded,
	})
	t.Require().NoError(err)

	err = t.repository.UpdateFileInfo(ctx, uuid.NewString(), repository2.UpdateFileInfoInput{
		Name:        "name-2",
		ContentType: "application/zip",
		Size:        100,
		Status:      repository2.FileStatusUploaded,
	})
	t.Require().ErrorIs(err, repository2.ErrNotFound)

	foundFile, err := t.repository.GetFile(ctx, file.ID)
	t.Require().NoError(err)
	t.Require().Equal(repository2.FileStatusUploaded, foundFile.Status)
	t.Require().NotNil(foundFile.Name)
	t.Require().Equal("name-1", *foundFile.Name)

	foundFile, err = t.repository.GetFileByName(ctx, "unknown")
	t.Require().ErrorIs(err, repository2.ErrNotFound)
	t.Require().Nil(foundFile)
}

func (t *testSuite) TestCreateFileParts() {
	ctx := context.Background()
	file := repository2.NewFile()

	err := t.repository.CreateFile(ctx, &file)
	t.Require().NoError(err)

	storage := repository2.NewStorage(uuid.NewString(), "127.0.0.1:9999")
	err = t.repository.CreateOrUpdateStorage(ctx, &storage)
	t.Require().NoError(err)

	for i := 0; i < 10; i++ {
		filePart := repository2.NewFilePart(file.ID, "", i, 100, storage.ID, uuid.NewString())
		err = t.repository.CreateFilePart(ctx, &filePart)
		t.Require().NoError(err)
	}

	foundFileParts, err := t.repository.FindOrderedFileParts(ctx, file.ID)
	t.Require().NoError(err)
	t.Require().Len(foundFileParts, 10)
}

func (t *testSuite) TestCreateStorage() {
	ctx := context.Background()

	storage := repository2.NewStorage(uuid.NewString(), "127.0.0.1:9999")
	err := t.repository.CreateOrUpdateStorage(ctx, &storage)
	t.Require().NoError(err)

	foundStorages, err := t.repository.FindStorages(ctx)
	t.Require().NoError(err)
	t.Require().Len(foundStorages, 1)
	t.Require().Equal("127.0.0.1:9999", foundStorages[0].Host)

	storage.Host = "127.0.0.1:8080"
	err = t.repository.CreateOrUpdateStorage(ctx, &storage)
	t.Require().NoError(err)

	foundStorages, err = t.repository.FindStorages(ctx)
	t.Require().NoError(err)
	t.Require().Len(foundStorages, 1)
	t.Require().Equal("127.0.0.1:8080", foundStorages[0].Host)
}
