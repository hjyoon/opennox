package blobs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatAccesses(t *testing.T) {
	blobTestFixture(t)
	const source = "package fixture\nfunc access() { _ = memmap.PtrUint32(0x1000, 2 + 2) }\n"
	require.NoError(t, os.WriteFile(Path("access.go"), []byte(source), 0644))
	err := FormatAccesses()
	require.NoError(t, err)
	got, err := os.ReadFile(Path("access.go"))
	require.NoError(t, err)
	require.Contains(t, string(got), "memmap.PtrUint32(0x1000, 4)")
}
