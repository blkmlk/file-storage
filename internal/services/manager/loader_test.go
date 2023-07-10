package manager

import (
	"bytes"
	"context"
	"crypto/rand"
	"testing"

	"github.com/blkmlk/file-storage/protocol"

	"github.com/stretchr/testify/require"

	"github.com/google/uuid"

	"github.com/blkmlk/file-storage/internal/mocks"
)

func TestLoader_Upload(t *testing.T) {
	ctx := context.Background()

	fullSize := int64(1792)
	ldr := NewLoader(fullSize)

	fileParts := []*FilePart{
		{
			StorageID: uuid.NewString(),
			Client:    mocks.NewStorage(ctx),
			Size:      597,
		},
		{
			StorageID: uuid.NewString(),
			Client:    mocks.NewStorage(ctx),
			Size:      597,
		},
		{
			StorageID: uuid.NewString(),
			Client:    mocks.NewStorage(ctx),
			Size:      598,
		},
	}
	for _, fp := range fileParts {
		resp, err := fp.Client.CheckReadiness(ctx, &protocol.CheckReadinessRequest{
			Size: fp.Size,
		})
		require.NoError(t, err)
		require.True(t, resp.Ready)

		fp.RemoteID = resp.Id
		ldr.AddFilePart(fp)
	}

	buff := make([]byte, fullSize)
	_, err := rand.Read(buff)
	require.NoError(t, err)

	err = ldr.Upload(ctx, bytes.NewReader(buff))
	require.NoError(t, err)

	offset := 0
	for _, fp := range ldr.GetFileParts() {
		s := fp.Client.(*mocks.Storage)
		parts := s.GetFileParts()
		require.Len(t, parts, 1)
		require.Equal(t, parts[0].ID, fp.RemoteID)
		require.Equal(t, parts[0].Size, fp.Size)
		require.Equal(t, parts[0].Data.Bytes(), buff[offset:offset+int(fp.Size)])
		require.NotEmpty(t, fp.Hash)

		offset += int(fp.Size)
	}

	var recovered bytes.Buffer
	err = ldr.Download(ctx, &recovered)
	require.NoError(t, err)
	require.Equal(t, buff, recovered.Bytes())
}
