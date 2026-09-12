package legacy

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"
)

func TestMapgenRenameRecords503230OracleSemantics(t *testing.T) {
	alpha := mapgenRecordWire502B10([]byte("Alpha"), 0x3f800000, 0xc0200000, []byte{1, 2, 3})
	beta := mapgenRecordWire502B10([]byte("Beta"), 0x7fc01234, 1, nil)
	lower := mapgenRecordWire502B10([]byte("beta"), 4, 5, []byte{6})
	withNUL := mapgenRecordWire502B10([]byte{'A', 0, 'B'}, 7, 8, []byte{9})
	for _, tc := range []struct {
		name    string
		oldName string
		newName string
		input   []byte
		want    []byte
	}{
		{"longer name, case-sensitive", "Beta", "Longer", mapgenStreamWire502B10(alpha, beta, lower), mapgenStreamWire502B10(alpha, mapgenRecordWire502B10([]byte("Longer"), 0x7fc01234, 1, nil), lower)},
		{"shorter name, all matches", "Beta", "B", mapgenStreamWire502B10(beta, alpha, beta), mapgenStreamWire502B10(mapgenRecordWire502B10([]byte("B"), 0x7fc01234, 1, nil), alpha, mapgenRecordWire502B10([]byte("B"), 0x7fc01234, 1, nil))},
		{"absent name", "Missing", "New", mapgenStreamWire502B10(alpha, beta), mapgenStreamWire502B10(alpha, beta)},
		{"empty stream", "Beta", "New", mapgenStreamWire502B10(), mapgenStreamWire502B10()},
		{"embedded NUL C comparison", "A", "Renamed", mapgenStreamWire502B10(withNUL, alpha), mapgenStreamWire502B10(mapgenRecordWire502B10([]byte("Renamed"), 7, 8, []byte{9}), alpha)},
		{"empty replacement", "Beta", "", mapgenStreamWire502B10(beta), mapgenStreamWire502B10(mapgenRecordWire502B10(nil, 0x7fc01234, 1, nil))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			if err := mapgenRenameRecords503230(bytes.NewReader(tc.input), &output, tc.oldName, tc.newName); err != nil || !bytes.Equal(output.Bytes(), tc.want) {
				t.Fatalf("rename = %v, %x; want %x", err, output.Bytes(), tc.want)
			}
		})
	}
}

func TestMapgenRenameRecords503230RejectsBadInputAndOutput(t *testing.T) {
	valid := mapgenStreamWire502B10(mapgenRecordWire502B10([]byte("Alpha"), 1, 2, nil))
	badMagic := bytes.Clone(valid)
	badMagic[0] = 0
	negativeLength := bytes.Clone(valid)
	binary.LittleEndian.PutUint32(negativeLength[4:8], math.MaxUint32)
	shortPayload := valid[:len(valid)-5]
	longName := mapgenStreamWire502B10(mapgenRecordWire502B10(bytes.Repeat([]byte{'N'}, 64), 1, 2, nil))
	for _, wire := range [][]byte{nil, badMagic, negativeLength, shortPayload, longName} {
		var output bytes.Buffer
		if err := mapgenRenameRecords503230(bytes.NewReader(wire), &output, "Alpha", "New"); err == nil {
			t.Fatalf("accepted malformed wire %x", wire)
		}
	}
	if err := mapgenRenameRecords503230(bytes.NewReader(valid), io.Discard, "Alpha", strings.Repeat("N", 64)); err == nil {
		t.Fatal("accepted replacement longer than reader record-name capacity")
	}
	if err := mapgenRenameRecords503230(bytes.NewReader(valid), mapgenShortWriter502ED0{}, "Alpha", "New"); !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("short write error = %v", err)
	}
}

func TestMapgenRenameAreaMap503230ViaC(t *testing.T) {
	initMapgenReaderHandles502B10(t)
	directory := t.TempDir()
	source := filepath.Join(directory, "map-index.tmp")
	backup := filepath.Join(directory, "AreaMap.bak")
	alpha := mapgenRecordWire502B10([]byte("Alpha"), 1, 2, []byte{8, 9})
	beta := mapgenRecordWire502B10([]byte("Beta"), 3, 4, nil)
	original := mapgenStreamWire502B10(alpha, beta, beta)
	if err := os.WriteFile(source, original, 0o600); err != nil {
		t.Fatal(err)
	}
	got := mapgenRenameViaC503230(source, directory, "Beta", "Longer")
	if got.result != 1 || got.count != 3 || !got.fileClosed || got.firstName != "Alpha" || got.secondName != "Longer" || got.thirdName != "Longer" {
		t.Fatalf("C rename and reload: %+v", got)
	}
	if unsafe.Sizeof(uintptr(0)) > 4 && got.recordsAddress <= math.MaxUint32 {
		t.Fatalf("record table pointer %#x was truncated", got.recordsAddress)
	}
	renamed := mapgenRecordWire502B10([]byte("Longer"), 3, 4, nil)
	want := mapgenStreamWire502B10(alpha, renamed, renamed)
	if data, err := os.ReadFile(source); err != nil || !bytes.Equal(data, want) {
		t.Fatalf("source = %x, %v; want %x", data, err, want)
	}
	if data, err := os.ReadFile(backup); err != nil || !bytes.Equal(data, original) {
		t.Fatalf("backup = %x, %v; want %x", data, err, original)
	}

	got = mapgenRenameViaC503230(source, directory, "Missing", "New")
	if got.result != 1 || got.count != 3 || got.secondName != "Longer" {
		t.Fatalf("absent name: %+v", got)
	}
	if data, err := os.ReadFile(source); err != nil || !bytes.Equal(data, want) {
		t.Fatalf("unchanged source = %x, %v", data, err)
	}
}

func TestMapgenRenameAreaMap503230EmptyAndRollback(t *testing.T) {
	initMapgenReaderHandles502B10(t)
	directory := t.TempDir()
	source := filepath.Join(directory, "map-index.tmp")
	backup := filepath.Join(directory, "AreaMap.bak")
	empty := mapgenStreamWire502B10()
	if err := os.WriteFile(source, empty, 0o600); err != nil {
		t.Fatal(err)
	}
	got := mapgenRenameViaC503230(source, directory, "Old", "New")
	if got.result != 1 || got.count != 0 || !got.fileClosed {
		t.Fatalf("empty stream: %+v", got)
	}
	if data, err := os.ReadFile(source); err != nil || !bytes.Equal(data, empty) {
		t.Fatalf("empty source = %x, %v", data, err)
	}

	valid := mapgenStreamWire502B10(mapgenRecordWire502B10([]byte("Old"), 1, 2, nil))
	for _, tc := range []struct {
		name string
		wire []byte
		new  string
	}{
		{"malformed wire", valid[:len(valid)-5], "New"},
		{"long replacement", valid, strings.Repeat("N", 64)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := os.WriteFile(source, tc.wire, 0o600); err != nil {
				t.Fatal(err)
			}
			got := mapgenRenameViaC503230(source, directory, "Old", tc.new)
			if got.result != 0 || !got.fileClosed {
				t.Fatalf("failed rename: %+v", got)
			}
			if data, err := os.ReadFile(source); err != nil || !bytes.Equal(data, tc.wire) {
				t.Fatalf("rollback source = %x, %v; want %x", data, err, tc.wire)
			}
			if _, err := os.Stat(backup); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("backup after rollback: %v", err)
			}
		})
	}
}
