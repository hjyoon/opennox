package blobs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// blobTestFixture keeps the legacy blob rewrite tools away from the source
// tree. The real blob data has since moved to generated .dat files, but these
// files exercise the parser, writer, and access rewriter's original format.
func blobTestFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"memmap.go": `package fixture
/*
extern unsigned char byte_1000[4];
extern unsigned char byte_2000[4];
extern unsigned char byte_3000[4];
*/
import "C"
import "unsafe"
func init() {
	memmap.RegisterBlobData(0x1000, "byte_1000", unsafe.Slice((*byte)(unsafe.Pointer(&C.byte_1000[0])), int(unsafe.Sizeof(C.byte_1000))))
	memmap.RegisterBlobData(0x2000, "byte_2000", unsafe.Slice((*byte)(unsafe.Pointer(&C.byte_2000[0])), int(unsafe.Sizeof(C.byte_2000))))
	memmap.RegisterBlobData(0x3000, "byte_3000", unsafe.Slice((*byte)(unsafe.Pointer(&C.byte_3000[0])), int(unsafe.Sizeof(C.byte_3000))))
}
`,
		"memmap.c": `extern unsigned char byte_1000[4];
extern unsigned char byte_2000[4];
extern unsigned char byte_3000[4];
void *mem_getPtrSize(unsigned base, unsigned off, unsigned size) {
	switch (base) {
	case 0x1000:
		if (off+size <= sizeof(byte_1000))
			return &byte_1000[off];
	case 0x2000:
		if (off+size <= sizeof(byte_2000))
			return &byte_2000[off];
	case 0x3000:
		if (off+size <= sizeof(byte_3000))
			return &byte_3000[off];
	}
	return 0;
}
`,
		"common/memmap/nox/noxmap.go": `package fixture
func init() {
	memmap.RegisterBlob(0x1000, "byte_1000", 4)
	memmap.RegisterBlob(0x2000, "byte_2000", 4)
	memmap.RegisterBlob(0x3000, "byte_3000", 4)
}
`,
		"GAME_data.c": `unsigned char byte_1000[4] = {1, 2, 3, 4};
unsigned char byte_2000[4] = {5, 6, 7, 8};
unsigned char byte_3000[4] = {9, 10, 11, 12};
`,
	}
	for name, data := range files {
		path := filepath.Join(dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
		require.NoError(t, os.WriteFile(path, []byte(data), 0644))
	}
	previous := Path()
	SetPath(dir)
	t.Cleanup(func() { SetPath(previous) })
	return dir
}

func TestReadBlobs(t *testing.T) {
	blobTestFixture(t)
	bl, err := ReadBlobs()
	require.NoError(t, err)
	const blobCnt = 3
	require.GreaterOrEqual(t, len(bl.data.blobs), blobCnt)
	require.GreaterOrEqual(t, len(bl.mapc.blobs1), blobCnt)
	require.GreaterOrEqual(t, len(bl.mapc.blobs2), blobCnt)
	require.GreaterOrEqual(t, len(bl.mapg1.blobs1), blobCnt)
	require.GreaterOrEqual(t, len(bl.mapg1.blobs2), blobCnt)
	require.GreaterOrEqual(t, len(bl.mapg2.blobs), blobCnt)
	err = bl.Write()
	require.NoError(t, err)
	bl, err = ReadBlobs()
	require.NoError(t, err)
	require.Len(t, bl.data.blobs, blobCnt)
	require.Len(t, bl.mapc.blobs1, blobCnt)
	require.Len(t, bl.mapc.blobs2, blobCnt)
	require.Len(t, bl.mapg1.blobs1, blobCnt)
	require.Len(t, bl.mapg1.blobs2, blobCnt)
	require.Len(t, bl.mapg2.blobs, blobCnt)
}
