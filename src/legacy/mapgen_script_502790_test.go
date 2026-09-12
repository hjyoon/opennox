package legacy

import (
	"encoding/binary"
	"io"
	"math"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

func TestMapgenScript502790OriginalMetadataAndWideHandle(t *testing.T) {
	InitBlobData()
	meta := mapScriptMetadata502790{}
	if got := meta.ArgCount(0); got != 1 {
		t.Fatalf("function zero argument count = %d, want 1", got)
	}
	if got := meta.ArgKind(0, 0); got != 1 {
		t.Fatalf("function zero argument kind = %d, want 1", got)
	}
	if got := meta.ArgCount(36); got != 0 {
		t.Fatalf("out-of-range function count = %d, want 0", got)
	}

	handles.Init()
	t.Cleanup(handles.Release)
	const nameLen = 1025
	wire := make([]byte, 4+nameLen+4+4+2+8+1)
	binary.LittleEndian.PutUint32(wire[0:4], nameLen)
	binary.LittleEndian.PutUint32(wire[4+nameLen:8+nameLen], 0xa1b2c3d4)
	binary.LittleEndian.PutUint32(wire[8+nameLen:12+nameLen], 1)
	wire[12+nameLen] = 0 // function zero: one eight-byte argument
	wire[len(wire)-1] = 0xee
	path := filepath.Join(t.TempDir(), "map-script.bin")
	if err := os.WriteFile(path, wire, 0o600); err != nil {
		t.Fatal(err)
	}
	cf, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	handle := NewFileHandle(cf.File.File)
	t.Cleanup(func() { nox_fs_close(handle) })
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(handle)) <= math.MaxUint32 {
		t.Fatalf("file handle = %p, want native-width address", handle)
	}
	handler := &server.ScriptCallback{Flags: 0x11223344, Func: 77}
	if got := mapgenMakeScript502790(unsafe.Pointer(handle), handler); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if *handler != (server.ScriptCallback{Flags: 0xa1b2c3d4, Func: 77}) {
		t.Fatalf("handler = %+v", *handler)
	}
	if off, err := cf.File.Seek(0, io.SeekCurrent); err != nil || off != int64(len(wire)-1) {
		t.Fatalf("cursor = %d, err = %v, want %d", off, err, len(wire)-1)
	}
}
