package legacy

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

func initMapgenReaderHandles502B10(t *testing.T) {
	t.Helper()
	handles.Init()
	t.Cleanup(handles.Release)
}

func mapgenRecordWire502B10(name []byte, xBits, yBits uint32, extra []byte) []byte {
	length := 11 + len(name) + len(extra)
	out := binary.LittleEndian.AppendUint32(nil, uint32(length))
	out = append(out, byte(len(name)))
	out = append(out, name...)
	out = append(out, 7, 2) // two bytes consumed but not retained by the index
	out = binary.LittleEndian.AppendUint32(out, xBits)
	out = binary.LittleEndian.AppendUint32(out, yBits)
	return append(out, extra...)
}

func mapgenStreamWire502B10(records ...[]byte) []byte {
	out := binary.LittleEndian.AppendUint32(nil, 0xCAFEDEAD)
	for _, record := range records {
		out = append(out, record...)
	}
	return binary.LittleEndian.AppendUint32(out, 0)
}

func TestMapgenReader502B10LazyAllocAndNativePointers(t *testing.T) {
	first, second, records, ok := mapgenReaderLazyAlloc502B10()
	if !ok || first == second || first == records || second == records {
		t.Fatalf("lazy mapgen allocation: first=%#x second=%#x records=%#x ok=%v", first, second, records, ok)
	}
	if unsafe.Sizeof(uintptr(0)) > 4 && (first <= math.MaxUint32 || second <= math.MaxUint32 || records <= math.MaxUint32) {
		t.Fatalf("mapgen allocations were not native-width: %#x %#x %#x", first, second, records)
	}
}

func TestMapgenReader502B10ValidWire(t *testing.T) {
	initMapgenReaderHandles502B10(t)
	first := mapgenRecordWire502B10([]byte("Alpha"), 0x3f800000, 0xc0200000, []byte{1, 2, 3})
	second := mapgenRecordWire502B10([]byte("Beta"), 0x7fc01234, 0x00000001, nil)
	path := filepath.Join(t.TempDir(), "map-names.bin")
	if err := os.WriteFile(path, mapgenStreamWire502B10(first, second), 0o600); err != nil {
		t.Fatal(err)
	}
	got := mapgenReadFile502B10(path)
	if got.result != 1 || got.count != 2 || !got.fileClosed {
		t.Fatalf("reader status: %+v", got)
	}
	if unsafe.Sizeof(uintptr(0)) > 4 && got.recordsAddress <= math.MaxUint32 {
		t.Fatalf("record table pointer %#x was truncated to PE32", got.recordsAddress)
	}
	if got.first != (mapgenReaderRecord502B10{"Alpha", 0x3f800000, 0xc0200000, 4}) {
		t.Fatalf("first record: %+v", got.first)
	}
	secondOffset := uint32(4 + len(first))
	if got.second != (mapgenReaderRecord502B10{"Beta", 0x7fc01234, 0x00000001, secondOffset}) || got.last != got.second {
		t.Fatalf("second/last records: %+v, %+v", got.second, got.last)
	}
}

func TestMapgenReader502B10NameBoundaryAndEmptyStream(t *testing.T) {
	initMapgenReaderHandles502B10(t)
	path := filepath.Join(t.TempDir(), "boundary.bin")
	if err := os.WriteFile(path, mapgenStreamWire502B10(), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := mapgenReadFile502B10(path); got.result != 1 || got.count != 0 || !got.fileClosed {
		t.Fatalf("empty stream: %+v", got)
	}

	name := bytes.Repeat([]byte{'Q'}, 63)
	wire := mapgenStreamWire502B10(mapgenRecordWire502B10(name, 0x7f800000, 0xff800000, nil))
	if err := os.WriteFile(path, wire, 0o600); err != nil {
		t.Fatal(err)
	}
	got := mapgenReadFile502B10(path)
	if got.result != 1 || got.count != 1 || got.first.name != string(name) || !got.fileClosed {
		t.Fatalf("63-byte name: %+v", got)
	}

	wire = mapgenStreamWire502B10(mapgenRecordWire502B10([]byte{'A', 0, 'B'}, 1, 2, nil))
	if err := os.WriteFile(path, wire, 0o600); err != nil {
		t.Fatal(err)
	}
	got = mapgenReadFile502B10(path)
	if got.result != 1 || got.count != 1 || got.first.name != "A" || !got.fileClosed {
		t.Fatalf("embedded NUL: %+v", got)
	}
}

func TestMapgenReader502B10RejectsMalformedWire(t *testing.T) {
	initMapgenReaderHandles502B10(t)
	valid := mapgenRecordWire502B10([]byte("Name"), 1, 2, nil)
	cases := []struct {
		name  string
		wire  []byte
		count uint32
	}{
		{"short magic", []byte{0xad, 0xde}, 0},
		{"wrong magic", []byte{0, 0, 0, 0}, 0},
		{"missing terminator", append(binary.LittleEndian.AppendUint32(nil, 0xCAFEDEAD), valid...), 1},
		{"short length", append(binary.LittleEndian.AppendUint32(nil, 0xCAFEDEAD), 0x01), 0},
		{"negative length", mapgenStreamWire502B10([]byte{0xff, 0xff, 0xff, 0xff}), 0},
		{"undersized record", mapgenStreamWire502B10([]byte{10, 0, 0, 0}), 0},
		{"name exceeds 64-byte slot", mapgenStreamWire502B10(mapgenRecordWire502B10(make([]byte, 64), 1, 2, nil)), 0},
		{"truncated name", append(binary.LittleEndian.AppendUint32(nil, 0xCAFEDEAD), 20, 0, 0, 0, 9, 'A'), 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "malformed.bin")
			if err := os.WriteFile(path, tc.wire, 0o600); err != nil {
				t.Fatal(err)
			}
			got := mapgenReadFile502B10(path)
			if got.result != 0 || got.count != tc.count || !got.fileClosed {
				t.Fatalf("reader status: %+v", got)
			}
		})
	}
}

func TestMapgenReader502B10EntryLimit(t *testing.T) {
	initMapgenReaderHandles502B10(t)
	record := mapgenRecordWire502B10([]byte("N"), 1, 2, nil)
	for _, tc := range []struct {
		name   string
		count  int
		result int
	}{
		{"maximum", 2048, 1},
		{"over maximum", 2049, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			records := make([][]byte, tc.count)
			for i := range records {
				records[i] = record
			}
			path := filepath.Join(t.TempDir(), "entries.bin")
			if err := os.WriteFile(path, mapgenStreamWire502B10(records...), 0o600); err != nil {
				t.Fatal(err)
			}
			got := mapgenReadFile502B10(path)
			if got.result != tc.result || got.count != 2048 || !got.fileClosed || got.last.name != "N" {
				t.Fatalf("reader status: %+v", got)
			}
		})
	}
}

func TestMapgenReader502B10MissingFile(t *testing.T) {
	initMapgenReaderHandles502B10(t)
	got := mapgenReadFile502B10(filepath.Join(t.TempDir(), "absent.bin"))
	if got.result != 0 || got.count != 0 || !got.fileClosed {
		t.Fatalf("reader status: %+v", got)
	}
}
