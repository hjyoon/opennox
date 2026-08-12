package main

import (
	"bytes"
	"debug/pe"
	"strings"
	"testing"
)

func TestValidateCodeManifest(t *testing.T) {
	m := validCodeManifest()
	if err := validateCodeManifest(&m); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		edit func(*codeManifest)
		want string
	}{
		{
			name: "noncanonical address",
			edit: func(m *codeManifest) { m.Functions[0].Address = "0x004e44f0" },
			want: "canonical",
		},
		{
			name: "overlap",
			edit: func(m *codeManifest) { m.Functions[1].Address = "0x004E44F8" },
			want: "sorted and disjoint",
		},
		{
			name: "bad digest",
			edit: func(m *codeManifest) { m.Functions[0].SHA256 = "00" },
			want: "invalid SHA-256",
		},
		{
			name: "duplicate name",
			edit: func(m *codeManifest) { m.Functions[1].Name = m.Functions[0].Name },
			want: "duplicate function name",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := validCodeManifest()
			tc.edit(&m)
			if err := validateCodeManifest(&m); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("validateCodeManifest() = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestReadCodeRange(t *testing.T) {
	data := []byte{0x90, 0x8b, 0x44, 0x24, 0x04, 0xc3}
	f := &pe.File{
		FileHeader:     pe.FileHeader{Machine: pe.IMAGE_FILE_MACHINE_I386},
		OptionalHeader: &pe.OptionalHeader32{Magic: 0x10b, ImageBase: 0x00400000},
		Sections: []*pe.Section{{
			SectionHeader: pe.SectionHeader{
				Name:            ".text",
				VirtualAddress:  0x1000,
				VirtualSize:     uint32(len(data)),
				Size:            uint32(len(data)),
				Characteristics: pe.IMAGE_SCN_MEM_EXECUTE,
			},
			ReaderAt: bytes.NewReader(data),
		}},
	}
	got, section, err := readCodeRange(f, 0x00401001, 4)
	if err != nil {
		t.Fatal(err)
	}
	if section != ".text" || !bytes.Equal(got, data[1:5]) {
		t.Fatalf("readCodeRange() = %x in %q", got, section)
	}
	if _, _, err := readCodeRange(f, 0x00401004, 4); err == nil || !strings.Contains(err.Error(), "outside") {
		t.Fatalf("out-of-range read error = %v", err)
	}
	f.Sections[0].Characteristics = 0
	if _, _, err := readCodeRange(f, 0x00401001, 1); err == nil || !strings.Contains(err.Error(), "not in an executable") {
		t.Fatalf("non-executable read error = %v", err)
	}
}

func validCodeManifest() codeManifest {
	return codeManifest{
		Schema:     codeManifestSchema,
		OracleID:   "test-copy",
		Path:       "GAME.EXE",
		FileSHA256: strings.Repeat("0", 64),
		Format:     "PE32/I386",
		ImageBase:  "0x00400000",
		Functions: []codeRecord{
			{Name: "first", Address: "0x004E44F0", Size: 15, SHA256: strings.Repeat("1", 64), Purpose: "first test"},
			{Name: "second", Address: "0x004E4500", Size: 94, SHA256: strings.Repeat("2", 64), Purpose: "second test"},
		},
	}
}
