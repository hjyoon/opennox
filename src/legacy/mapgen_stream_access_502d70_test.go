package legacy

import (
	"math"
	"os"
	"path/filepath"
	"testing"
	"unsafe"
)

func TestMapgenStreamAccess502D70NativePointersAndLookup(t *testing.T) {
	initMapgenReaderHandles502B10(t)
	path := filepath.Join(t.TempDir(), "stream.bin")
	if err := os.WriteFile(path, make([]byte, 32), 0o600); err != nil {
		t.Fatal(err)
	}
	got := mapgenStreamAccess502D70(path, filepath.Join(t.TempDir(), "absent.bin"))
	if unsafe.Sizeof(uintptr(0)) > 4 && got.recordsAddress <= math.MaxUint32 {
		t.Fatalf("record pointer %#x was truncated to PE32", got.recordsAddress)
	}
	if got.fileAddress == 0 || got.x != 1.5 || got.y != -2.5 ||
		!got.invalidIndex || !got.invalidCoordinate || !got.closedSeek || !got.opened ||
		!got.seekByIndex || !got.seekByName || !got.absentName || !got.nilName ||
		!got.reopenRewinds || !got.closed || !got.missingFile || !got.idempotentClose {
		t.Fatalf("map-name stream access: %+v", got)
	}
}
