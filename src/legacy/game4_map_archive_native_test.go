package legacy

import (
	"bytes"
	"errors"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"unsafe"
)

func TestGame4MapArchiveNativeOwnersAndGC(t *testing.T) {
	b := game4ArchiveBankNew()
	if b == nil {
		t.Fatal("C-owned fixture allocation failed")
	}
	defer b.close()
	b.seed()
	want := b.leaf(true)
	wantBytes := b.payload()
	b.seed()
	got := b.leaf(false)
	runtime.GC()
	if got != want || !bytes.Equal(b.payload(), wantBytes) {
		t.Fatalf("public native C entries differ from retained C reference: got %+v, want %+v", got, want)
	}
	base := b.address()
	names := b.nameAddresses()
	if unsafe.Sizeof(uintptr(0)) == 8 && base <= math.MaxUint32 {
		t.Fatalf("C record owner is not high-address: %#x", base)
	}
	if got.pointers != [4]uintptr{base + 2*76, base + 3*76, names[0], names[1]} || names[0] == 0 || names[1] == 0 || names[0] == names[1] {
		t.Fatalf("borrowed pointer identities differ: %x (records=%#x)", got.pointers, base)
	}
	if !b.ownersIntact() {
		t.Fatal("native C owner slots or adjacent canaries changed")
	}
	// Call through the public C ABI after collection while the same C bank
	// remains installed; no Go pointer was retained in a C global owner.
	if again := b.leaf(false); again != got {
		t.Fatalf("public entry results changed across GC: %+v -> %+v", got, again)
	}
}

func TestGame4MapArchiveIndexCurrentCReference(t *testing.T) {
	initMapgenReaderHandles502B10(t)
	b := game4ArchiveBankNew()
	if b == nil {
		t.Fatal("C-owned fixture allocation failed")
	}
	defer b.close()
	path := filepath.Join(t.TempDir(), "index.bin")
	valid := mapgenStreamWire502B10(
		mapgenRecordWire502B10([]byte{'A', 0, 'Z'}, 0x7fc01234, 0x80000000, bytes.Repeat([]byte{1, 2, 3}, 1500)),
		mapgenRecordWire502B10(bytes.Repeat([]byte{'Q'}, 63), 0x00000001, 0xff800000, nil),
	)
	for _, tc := range []struct {
		name string
		wire []byte
	}{
		{"valid raw fields and embedded NUL", valid},
		{"truncated second record", valid[:len(valid)-9]},
		{"invalid magic", []byte{0, 0, 0, 0}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := os.WriteFile(path, tc.wire, 0o600); err != nil {
				t.Fatal(err)
			}
			b.seed()
			wr, wc, wclosed := b.index(path, true)
			want := b.payload()
			b.seed()
			gr, gc, gclosed := b.index(path, false)
			if gr != wr || gc != wc || gclosed != wclosed || !gclosed || !bytes.Equal(b.payload(), want) || !b.ownersIntact() {
				t.Fatalf("index/reference mismatch: got(%d,%d,%v), want(%d,%d,%v)", gr, gc, gclosed, wr, wc, wclosed)
			}
		})
	}
}

func TestGame4MapArchiveGeometryCurrentCReference(t *testing.T) {
	for _, tc := range []struct {
		name       string
		x, y, w, h float32
		ox, oy     int32
		wall       uint32
	}{
		{"ordinary", 100, 200, 46, 92, 3000, 3000, 20},
		{"clamped wrapped", -10000, -10000, 46, 92, math.MinInt32, math.MaxInt32, 0x40000000},
		{"negative dimensions", 2400, 810, -46, -92, -1, 1, 0xffffffff},
		{"coincident corners", 0, 0, 0, 0, 0, 0, 0},
		{"float boundary", 2260.84130859375, 810.8399658203125, 0.125, -0.125, math.MaxInt32, math.MinInt32, 0x80000000},
		{"unordered", float32(math.NaN()), 12, float32(math.Inf(1)), 0, 0, 0, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want := game4ArchiveGeometryReference(tc.x, tc.y, tc.w, tc.h, tc.ox, tc.oy, tc.wall, true)
			got := game4ArchiveGeometryReference(tc.x, tc.y, tc.w, tc.h, tc.ox, tc.oy, tc.wall, false)
			if got != want || !got.canaries {
				t.Fatalf("geometry differs: got %+v, want retained C %+v", got, want)
			}
		})
	}
	for _, tc := range []struct {
		name       string
		coordinate float32
		origin     int32
	}{
		{"offset low signed wrap", 0x1p31, 0},
		{"offset origin subtraction", 0x1p32, math.MinInt32},
		{"offset qword positive invalid", 0x1p63, 0},
		{"offset qword negative edge", -0x1p63, 0},
		{"offset NaN", float32(math.NaN()), 0},
		{"offset infinity", float32(math.Inf(-1)), math.MaxInt32},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want := game4ArchiveOffsetReference(tc.coordinate, tc.origin, true)
			if got := game4ArchiveOffsetReference(tc.coordinate, tc.origin, false); got != want {
				t.Fatalf("offset differs: %d / retained C %d", got, want)
			}
		})
	}
}

func TestGame4MapArchiveExtractCurrentCReference(t *testing.T) {
	initMapgenReaderHandles502B10(t)
	b := game4ArchiveBankNew()
	if b == nil {
		t.Fatal("C-owned fixture allocation failed")
	}
	defer b.close()
	dir := t.TempDir()
	source := filepath.Join(dir, "AreaMap.dat")
	attachment := filepath.Join(dir, "embedded.bin")
	payload := bytes.Repeat([]byte{7, 8, 0, 9}, 2300)
	optional := bytes.Repeat([]byte{1, 2, 0, 3}, 1100)
	read := func(path string) ([]byte, bool) {
		data, err := os.ReadFile(path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			t.Fatal(err)
		}
		return data, err == nil
	}
	for _, tc := range []struct {
		name    string
		version byte
		lookup  string
	}{
		{"map and attachment", 2, "tArGeT"},
		{"map without attachment", 1, "Target"},
		{"missing record deletes stale attachment", 2, "Absent"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			wire := mapgenStreamWire502B10(mapgenExtractRecord5034B0("Target", tc.version, optional, payload))
			if err := os.WriteFile(source, wire, 0o600); err != nil {
				t.Fatal(err)
			}
			var wantResult int
			var wantMap, wantAttachment []byte
			var wantMapExists, wantAttachmentExists bool
			for _, reference := range []bool{true, false} {
				if err := os.Remove(filepath.Join(dir, "copy.tmp")); err != nil && !errors.Is(err, os.ErrNotExist) {
					t.Fatal(err)
				}
				if err := os.WriteFile(attachment, []byte("stale owned output"), 0o600); err != nil {
					t.Fatal(err)
				}
				b.seed()
				got := b.extract(source, dir, tc.lookup, attachment, reference)
				mapBytes, mapExists := read(filepath.Join(dir, "copy.tmp"))
				attachmentBytes, attachmentExists := read(attachment)
				if !b.ownersIntact() {
					t.Fatal("extraction changed native owner slots/canaries")
				}
				if reference {
					wantResult = got
					wantMap = mapBytes
					wantAttachment = attachmentBytes
					wantMapExists = mapExists
					wantAttachmentExists = attachmentExists
					continue
				}
				if got != wantResult || mapExists != wantMapExists || attachmentExists != wantAttachmentExists || !bytes.Equal(mapBytes, wantMap) || !bytes.Equal(attachmentBytes, wantAttachment) {
					t.Fatalf("public extraction differs from retained C: result %d/%d map %v/%v attachment %v/%v", got, wantResult, mapExists, wantMapExists, attachmentExists, wantAttachmentExists)
				}
			}
			if after, err := os.ReadFile(source); err != nil || !bytes.Equal(after, wire) {
				t.Fatalf("extraction changed input: %v", err)
			}
		})
	}
}
