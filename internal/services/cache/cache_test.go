package cache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMapCache(t *testing.T) {
	mc := NewMapCache()

	keys := []string{"key1", "key2"}
	err := mc.Lock(keys)
	require.NoError(t, err)

	err = mc.Lock([]string{"key1"})
	require.Error(t, err)

	err = mc.Lock([]string{"key2"})
	require.Error(t, err)

	err = mc.Lock(keys)
	require.Error(t, err)

	mc.Unlock(keys)
	err = mc.Lock(keys)
	require.NoError(t, err)
}
