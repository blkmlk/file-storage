package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/blkmlk/file-storage/deps"
	"github.com/blkmlk/file-storage/migrations"
	"github.com/blkmlk/file-storage/services/repository"
	"github.com/stretchr/testify/suite"
	"go.uber.org/dig"
)

func TestAll(t *testing.T) {
	suite.Run(t, new(testSuite))
}

type testSuite struct {
	suite.Suite
	ctn        *dig.Container
	repository repository.Repository
}

func (t *testSuite) SetupTest() {
	t.ctn = dig.New()
	t.Require().NoError(t.ctn.Provide(deps.NewLocalDB))
	t.Require().NoError(t.ctn.Provide(repository.New))

	m, err := migrations.NewLocal()
	t.Require().NoError(err)

	if m.Up() != nil {
		t.Require().NoError(m.Drop())
		m, _ = migrations.NewLocal()
		t.Require().NoError(m.Up())
	}

	err = t.ctn.Invoke(func(repo repository.Repository) {
		t.repository = repo
	})
	t.Require().NoError(err)
}

func (t *testSuite) TestCreateUpdateAndGetFile() {
	ctx := context.Background()
	file := repository.NewFile("test")

	err := t.repository.CreateFile(ctx, &file)
	t.Require().NoError(err)

	err = t.repository.CreateFile(ctx, &file)
	t.Require().ErrorIs(err, repository.ErrAlreadyExists)

	err = t.repository.UpdateFileStatus(ctx, file.ID, "hash", repository.FileStatusUploaded)
	t.Require().NoError(err)

	err = t.repository.UpdateFileStatus(ctx, uuid.NewString(), "hash", repository.FileStatusUploaded)
	t.Require().ErrorIs(err, repository.ErrNotFound)

	foundFile, err := t.repository.GetFile(ctx, file.Name)
	t.Require().NoError(err)
	t.Require().Equal(repository.FileStatusUploaded, foundFile.Status)

	foundFile, err = t.repository.GetFile(ctx, "unknown")
	t.Require().ErrorIs(err, repository.ErrNotFound)
	t.Require().Nil(foundFile)
}

func (t *testSuite) TestCreateFileParts() {
	ctx := context.Background()
	file := repository.NewFile("test")

	err := t.repository.CreateFile(ctx, &file)
	t.Require().NoError(err)

	fileStorage := repository.NewFileStorage(uuid.NewString())
	err = t.repository.CreateFileStorage(ctx, &fileStorage)
	t.Require().NoError(err)

	for i := 0; i < 10; i++ {
		filePart := repository.NewFilePart(file.ID, i, fileStorage.ID, uuid.NewString())
		err = t.repository.CreateFilePart(ctx, &filePart)
		t.Require().NoError(err)
	}

	foundFileParts, err := t.repository.FindFileParts(ctx, file.ID)
	t.Require().NoError(err)
	t.Require().Len(foundFileParts, 10)
}
