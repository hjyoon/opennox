package legacy

import (
	"bytes"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

func TestMapgenThemeNativeEncryptedBinfileR310(t *testing.T) {
	// Unit tests do not run blobdata.Init. Use its exact immutable bytes and
	// two native side-slot assignments, preserving the old packed table.
	blob, err := os.ReadFile("../common/memmap/nox/blobdata/blob_587000.dat")
	if err != nil {
		t.Fatal(err)
	}
	for _, region := range []struct {
		base, offset uintptr
		size         int
	}{{0x587000, 253488, 18}, {0x5D4594, 2487264, 256}, {0x5D4594, 2487520, 4}} {
		data := unsafe.Slice((*byte)(memmap.PtrOff(region.base, region.offset)), region.size)
		old := bytes.Clone(data)
		t.Cleanup(func() { copy(data, old) })
	}
	packed := unsafe.Slice((*byte)(memmap.PtrOff(0x587000, 253244)), 12)
	packedBefore := bytes.Clone(packed)
	for i := 0; i < 3; i++ {
		slot := memmap.PtrPtr(0x587000, uintptr(253244+4*i))
		old := *slot
		t.Cleanup(func() { *slot = old })
	}
	copy(unsafe.Slice((*byte)(memmap.PtrOff(0x587000, 253488)), 18), blob[253488:253506])
	*memmap.PtrPtr(0x587000, 253244) = memmap.PtrOff(0x587000, 253488)
	*memmap.PtrPtr(0x587000, 253248) = memmap.PtrOff(0x587000, 253496)
	*memmap.PtrPtr(0x587000, 253252) = nil

	files.Lock()
	oldFiles := files.byHandle
	files.byHandle = make(map[unsafe.Pointer]*binfile.File)
	files.Unlock()
	handles.Init()
	t.Cleanup(func() {
		files.Lock()
		owned := files.byHandle
		files.byHandle = oldFiles
		files.Unlock()
		for _, file := range owned {
			_ = file.Close()
		}
		handles.Release()
	})
	t.Chdir(t.TempDir())
	if err := os.Mkdir("mapgen", 0o700); err != nil {
		t.Fatal(err)
	}
	const valid = `ALGORITHM_DATA
midHallLength
17
skeleton
HALL_RING
useDoors
TRUE
debug
false
mapSize
2.75
recursionLimit
9
END
AMBIENT_LIGHT
12
34
56
EXIT
OBJECT
ExitN
NORTH
LINKDATA
Gate42
END
SPELL_SET
END
DECOR
ROOM
RoomBase
END
DECOR
HALL
HallBase
END
`
	for _, tc := range []struct {
		name, fileName, contents string
		legacy, wantSuccess      bool
	}{
		{"native pointers", "native", valid, false, true},
		{"registered high-bit name token", "token", valid, true, true},
		{"original bounded path capacity", strings.Repeat("T", 244), valid, false, true},
		{"partial ambient EOF", "incomplete", "AMBIENT_LIGHT\n12\n34\n", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join("mapgen", tc.fileName+".thm")
			file, err := binfile.BinfileOpen(path, binfile.WriteOnly)
			if err != nil {
				t.Fatal(err)
			}
			if err := file.SetKey(1); err != nil {
				_ = file.Close()
				t.Fatal(err)
			}
			if n, err := file.Write([]byte(tc.contents)); err != nil || n != len(tc.contents) {
				_ = file.Close()
				t.Fatalf("write encrypted theme: n=%d error=%v", n, err)
			}
			if err := file.Close(); err != nil {
				t.Fatal(err)
			}
			got := mapgenThemeFixtureR310(tc.fileName, tc.legacy)
			if got.result != tc.wantSuccess || !got.guardsPreserved || !got.cleanupCleared || tc.legacy && strconv.IntSize == 64 && !got.highNameToken {
				t.Fatalf("native theme result=%+v", got)
			}
			addresses := map[string]uintptr{"theme": got.themeAddress, "name": got.nameAddress}
			if tc.wantSuccess {
				if got.words[0] != 1 || got.words[1] != 17 || got.words[14] != 1 || got.words[15] != 0 || got.words[16] != 0x40300000 || got.words[18] != 9 ||
					got.words[134] != 12 || got.words[135] != 34 || got.words[136] != 56 || got.words[118] != 1 || got.words[69] != 0 || got.words[274] != 0 ||
					got.words[23] != 1 || got.words[31] != 1 || got.roomName != "RoomBase" || got.hallName != "HallBase" || got.exitName != "ExitN" || got.linkData != "Gate42" {
					t.Fatalf("encrypted theme fields/graph changed: %+v", got)
				}
				for i := 0; i < 6; i++ {
					if got.words[24+i] != 1000 || got.words[32+i] != 1000 {
						t.Fatalf("original room/hall frequency aggregates at %d: %d/%d", i, got.words[24+i], got.words[32+i])
					}
				}
				addresses["room"], addresses["hall"] = got.roomAddress, got.hallAddress
			} else if got.words[134] != 12 || got.words[135] != 34 || got.words[136] != 0 || got.words[23] != 0 || got.words[31] != 0 {
				t.Fatalf("premature EOF lost its original partial writes: %+v", got)
			}
			files.RLock()
			found := 0
			for handle, opened := range files.byHandle {
				if filepath.Base(opened.Name()) != tc.fileName+".thm" {
					continue
				}
				found++
				addresses["actual binfile handle"] = uintptr(handle)
				if _, err := opened.Stat(); !errors.Is(err, os.ErrClosed) {
					files.RUnlock()
					t.Fatalf("theme reader did not close its actual encrypted binfile: %v", err)
				}
			}
			files.RUnlock()
			if found != 1 {
				t.Fatalf("actual native handle observations=%d, want one", found)
			}
			for name, address := range addresses {
				if address == 0 || strconv.IntSize == 64 && uint64(address) <= uint64(math.MaxUint32) {
					t.Fatalf("%s address=%#x does not exercise native pointer width", name, address)
				}
			}
			if strconv.IntSize == 64 && !bytes.Equal(packed, packedBefore) {
				t.Fatal("native skeleton lookup rewrote the packed PE32 table")
			}
		})
	}
}
