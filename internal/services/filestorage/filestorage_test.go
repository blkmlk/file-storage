package filestorage

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"os"
	"testing"

	"github.com/blkmlk/file-storage/env"
	"github.com/stretchr/testify/require"
)

func TestFSStorage(t *testing.T) {
	dirName, err := os.MkdirTemp("", "fs-test")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(dirName)
	}()

	_ = os.Setenv(env.FSRootPath, dirName)

	ctx := context.Background()

	fs, err := NewFSStorage()
	require.NoError(t, err)

	buff := make([]byte, 1024)
	_, err = rand.Read(buff)
	require.NoError(t, err)

	exists, err := fs.Exists(ctx, "test")
	require.NoError(t, err)
	require.False(t, exists)

	writer, err := fs.Create(ctx, "test")
	require.NoError(t, err)

	_, err = io.Copy(writer, bytes.NewReader(buff))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	exists, err = fs.Exists(ctx, "test")
	require.NoError(t, err)
	require.True(t, exists)

	reader, err := fs.Get(ctx, "test")
	require.NoError(t, err)

	fileBytes, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, buff, fileBytes)
}
