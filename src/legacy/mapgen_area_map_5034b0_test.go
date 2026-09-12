package legacy

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"
)

func mapgenExtractRecord5034B0(name string, version byte, attachment, payload []byte) []byte {
	extra := make([]byte, 0, 4+len(attachment)+4+len(payload))
	if version > 1 {
		extra = binary.LittleEndian.AppendUint32(extra, uint32(len(attachment)))
		extra = append(extra, attachment...)
	}
	extra = binary.LittleEndian.AppendUint32(extra, mapgenMagic502ED0)
	extra = append(extra, payload...)
	record := mapgenRecordWire502B10([]byte(name), 1, 2, extra)
	record[4+1+len(name)+1] = version
	return record
}

func TestMapgenExtract5034B0ViaC(t *testing.T) {
	initMapgenReaderHandles502B10(t)
	directory := t.TempDir()
	source := filepath.Join(directory, "AreaMap.dat")
	attachmentPath := filepath.Join(directory, "embedded.bin")
	attachment := []byte{0, 1, 2, 3, 0xff}
	payload := bytes.Repeat([]byte{7, 8, 9, 0, 10}, 1700)
	wire := mapgenStreamWire502B10(
		mapgenExtractRecord5034B0("First", 1, nil, []byte{1}),
		mapgenExtractRecord5034B0("Target", 2, attachment, payload),
	)
	if err := os.WriteFile(source, wire, 0o600); err != nil {
		t.Fatal(err)
	}
	got := mapgenExtractViaC5034B0(source, directory, "target", attachmentPath)
	if got.indexed != 1 || !got.payload || !got.attachment || !got.fileClosed {
		t.Fatalf("C extraction = %+v", got)
	}
	if unsafe.Sizeof(uintptr(0)) > 4 && got.recordsAddress <= math.MaxUint32 {
		t.Fatalf("AreaMap records pointer %#x was narrowed", got.recordsAddress)
	}
	wantCopy := binary.LittleEndian.AppendUint32(nil, 0xABEDFACE)
	wantCopy = append(wantCopy, payload...)
	if data, err := os.ReadFile(filepath.Join(directory, "copy.tmp")); err != nil || !bytes.Equal(data, wantCopy) {
		t.Fatalf("copy.tmp = %x, %v; want %x", data, err, wantCopy)
	}
	if data, err := os.ReadFile(attachmentPath); err != nil || !bytes.Equal(data, attachment) {
		t.Fatalf("attachment = %x, %v; want %x", data, err, attachment)
	}
}

func TestMapgenExtract5034B0FailurePaths(t *testing.T) {
	initMapgenReaderHandles502B10(t)
	directory := t.TempDir()
	source := filepath.Join(directory, "AreaMap.dat")
	attachmentPath := filepath.Join(directory, "embedded.bin")
	copyPath := filepath.Join(directory, "copy.tmp")
	for _, tc := range []struct {
		name           string
		record         []byte
		lookup         string
		wantPayload    bool
		wantAttachment bool
	}{
		{"no optional block", mapgenExtractRecord5034B0("Target", 1, nil, []byte{1, 2}), "Target", true, false},
		{"empty optional block", mapgenExtractRecord5034B0("Target", 2, nil, []byte{1, 2}), "Target", true, false},
		{"missing name", mapgenExtractRecord5034B0("Target", 2, []byte{3}, []byte{4}), "Absent", false, false},
		{"bad payload magic", func() []byte {
			record := mapgenExtractRecord5034B0("Target", 2, []byte{3}, []byte{4})
			record[len(record)-5] = 0
			return record
		}(), "Target", false, true},
		{"oversized optional block", func() []byte {
			record := mapgenExtractRecord5034B0("Target", 2, []byte{3}, []byte{4})
			binary.LittleEndian.PutUint32(record[15+len("Target"):], 100)
			return record
		}(), "Target", false, false},
		{"zero-length payload", mapgenExtractRecord5034B0("Target", 2, []byte{3}, nil), "Target", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := os.WriteFile(source, mapgenStreamWire502B10(tc.record), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(attachmentPath, []byte("stale"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(copyPath); err != nil && !errors.Is(err, os.ErrNotExist) {
				t.Fatal(err)
			}
			got := mapgenExtractViaC5034B0(source, directory, tc.lookup, attachmentPath)
			if got.indexed != 1 || got.payload != tc.wantPayload || got.attachment != tc.wantAttachment || !got.fileClosed {
				t.Fatalf("C extraction = %+v", got)
			}
			if !tc.wantAttachment {
				if _, err := os.Stat(attachmentPath); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("failed attachment extraction left stale file: %v", err)
				}
			}
			if !tc.wantPayload {
				if _, err := os.Stat(copyPath); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("failed payload extraction created copy.tmp: %v", err)
				}
			}
		})
	}
	valid := mapgenExtractRecord5034B0("Target", 2, []byte{3}, []byte{4})
	if err := os.WriteFile(source, mapgenStreamWire502B10(valid), 0o600); err != nil {
		t.Fatal(err)
	}
	got := mapgenExtractViaC5034B0(source, strings.Repeat("x", 2040), "Target", attachmentPath)
	if got.payload || !got.attachment || !got.fileClosed {
		t.Fatalf("overlong copy.tmp path = %+v", got)
	}
}
