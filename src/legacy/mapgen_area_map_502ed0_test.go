package legacy

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os"
	"path/filepath"
	"testing"
	"unsafe"
)

func TestMapgenRewriteRecords502ED0OracleSemantics(t *testing.T) {
	alpha := mapgenRecordWire502B10([]byte("Alpha"), 0x3f800000, 0xc0200000, []byte{1, 2, 3})
	beta := mapgenRecordWire502B10([]byte("Beta"), 0x7fc01234, 1, nil)
	lower := mapgenRecordWire502B10([]byte("beta"), 4, 5, []byte{6})
	withNUL := mapgenRecordWire502B10([]byte{'A', 0, 'B'}, 7, 8, nil)
	for _, tc := range []struct {
		name   string
		target string
		input  []byte
		want   []byte
		result bool
	}{
		{"middle, case-sensitive", "Beta", mapgenStreamWire502B10(alpha, beta, lower), mapgenStreamWire502B10(alpha, lower), true},
		{"absent name still returns true", "Missing", mapgenStreamWire502B10(alpha, beta), mapgenStreamWire502B10(alpha, beta), true},
		{"last record removed", "Beta", mapgenStreamWire502B10(beta), mapgenStreamWire502B10(), false},
		{"empty stream", "Beta", mapgenStreamWire502B10(), mapgenStreamWire502B10(), false},
		{"C string comparison", "A", mapgenStreamWire502B10(withNUL, alpha), mapgenStreamWire502B10(alpha), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			got, err := mapgenRewriteRecords502ED0(bytes.NewReader(tc.input), &output, tc.target)
			if err != nil || got != tc.result || !bytes.Equal(output.Bytes(), tc.want) {
				t.Fatalf("rewrite = %t, %v, %x; want %t, %x", got, err, output.Bytes(), tc.result, tc.want)
			}
		})
	}
}

type mapgenShortWriter502ED0 struct{}

func (mapgenShortWriter502ED0) Write(p []byte) (int, error) { return 0, nil }

func TestMapgenRewriteRecords502ED0RejectsBadInputAndOutput(t *testing.T) {
	valid := mapgenStreamWire502B10(mapgenRecordWire502B10([]byte("Alpha"), 1, 2, nil))
	badMagic := bytes.Clone(valid)
	badMagic[0] = 0
	negativeLength := bytes.Clone(valid)
	binary.LittleEndian.PutUint32(negativeLength[4:8], math.MaxUint32)
	shortPayload := valid[:len(valid)-5]
	longName := mapgenStreamWire502B10(mapgenRecordWire502B10(bytes.Repeat([]byte{'N'}, 64), 1, 2, nil))
	for _, wire := range [][]byte{nil, badMagic, negativeLength, shortPayload, longName} {
		var output bytes.Buffer
		if _, err := mapgenRewriteRecords502ED0(bytes.NewReader(wire), &output, "Missing"); err == nil {
			t.Fatalf("accepted malformed wire %x", wire)
		}
	}
	if _, err := mapgenRewriteRecords502ED0(bytes.NewReader(valid), mapgenShortWriter502ED0{}, "Missing"); !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("short write error = %v", err)
	}
}

func TestMapgenRewriteAreaMap502ED0ViaC(t *testing.T) {
	initMapgenReaderHandles502B10(t)
	directory := t.TempDir()
	source := filepath.Join(directory, "map-index.tmp")
	backup := filepath.Join(directory, "AreaMap.bak")
	alpha := mapgenRecordWire502B10([]byte("Alpha"), 1, 2, []byte{8, 9})
	beta := mapgenRecordWire502B10([]byte("Beta"), 3, 4, nil)
	gamma := mapgenRecordWire502B10([]byte("Gamma"), 5, 6, []byte{7})
	input := mapgenStreamWire502B10(alpha, beta, gamma)
	if err := os.WriteFile(source, input, 0o600); err != nil {
		t.Fatal(err)
	}
	got := mapgenRewriteViaC502ED0(source, directory, "Beta")
	if got.result != 1 || got.count != 2 || !got.fileClosed || got.firstName != "Alpha" || got.secondName != "Gamma" {
		t.Fatalf("C rewrite and reload: %+v", got)
	}
	if unsafe.Sizeof(uintptr(0)) > 4 && got.recordsAddress <= math.MaxUint32 {
		t.Fatalf("record table pointer %#x was truncated", got.recordsAddress)
	}
	want := mapgenStreamWire502B10(alpha, gamma)
	if data, err := os.ReadFile(source); err != nil || !bytes.Equal(data, want) {
		t.Fatalf("source = %x, %v; want %x", data, err, want)
	}
	if data, err := os.ReadFile(backup); err != nil || !bytes.Equal(data, input) {
		t.Fatalf("backup = %x, %v; want %x", data, err, input)
	}

	got = mapgenRewriteViaC502ED0(source, directory, "Missing")
	if got.result != 1 || got.count != 2 || got.firstName != "Alpha" || got.secondName != "Gamma" {
		t.Fatalf("absent name: %+v", got)
	}
	if data, err := os.ReadFile(source); err != nil || !bytes.Equal(data, want) {
		t.Fatalf("unchanged source = %x, %v", data, err)
	}
}

func TestMapgenRewriteAreaMap502ED0EmptyAndRollback(t *testing.T) {
	initMapgenReaderHandles502B10(t)
	directory := t.TempDir()
	source := filepath.Join(directory, "map-index.tmp")
	backup := filepath.Join(directory, "AreaMap.bak")
	only := mapgenStreamWire502B10(mapgenRecordWire502B10([]byte("Only"), 1, 2, nil))
	if err := os.WriteFile(source, only, 0o600); err != nil {
		t.Fatal(err)
	}
	got := mapgenRewriteViaC502ED0(source, directory, "Only")
	if got.result != 0 || got.count != 0 || !got.fileClosed {
		t.Fatalf("last record removed: %+v", got)
	}
	if data, err := os.ReadFile(source); err != nil || !bytes.Equal(data, mapgenStreamWire502B10()) {
		t.Fatalf("empty source = %x, %v", data, err)
	}
	if data, err := os.ReadFile(backup); err != nil || !bytes.Equal(data, only) {
		t.Fatalf("empty result backup = %x, %v", data, err)
	}

	malformed := only[:len(only)-5]
	if err := os.WriteFile(source, malformed, 0o600); err != nil {
		t.Fatal(err)
	}
	got = mapgenRewriteViaC502ED0(source, directory, "Absent")
	if got.result != 0 || !got.fileClosed {
		t.Fatalf("malformed rewrite: %+v", got)
	}
	if data, err := os.ReadFile(source); err != nil || !bytes.Equal(data, malformed) {
		t.Fatalf("rollback source = %x, %v; want %x", data, err, malformed)
	}
	if _, err := os.Stat(backup); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("backup after rollback: %v", err)
	}
}
