package blobs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSplitBlob(t *testing.T) {
	blobTestFixture(t)
	require.NoError(t, SplitBlob(0x1000, 2, 0))
	bl, err := ReadBlobs()
	require.NoError(t, err)
	left := bl.Get(0x1000)
	require.NotNil(t, left)
	require.Equal(t, uintptr(2), left.Size)
	require.Equal(t, []byte{1, 2}, left.Data)
	right := bl.Get(0x1002)
	require.NotNil(t, right)
	require.Equal(t, uintptr(2), right.Size)
	require.Equal(t, []byte{3, 4}, right.Data)
}
