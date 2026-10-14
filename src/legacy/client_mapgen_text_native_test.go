package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestClientMapGenText4A9D00NativeOriginalContracts(t *testing.T) {
	var original struct {
		SHA   string `json:"original_sha256"`
		Bytes int    `json:"code_bytes"`
		Code  string `json:"code_sha256"`
		Rows  []struct {
			Input  [10]int64  `json:"input"`
			Events [][5]int64 `json:"events"`
			After  [3]uint32  `json:"after"`
		} `json:"rows"`
	}
	data, err := os.ReadFile("testdata/mapgen_text_4a9d00_original.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &original); err != nil {
		t.Fatal(err)
	}
	if original.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || original.Bytes != 395 ||
		original.Code != "684af341bcc887d070e0ef90ed89c618b08b97681f85b0efbb956e90e642c0e9" || len(original.Rows) != 227 {
		t.Fatal("wrong sealed original map-generation entry")
	}
	var input bytes.Buffer
	if err := binary.Write(&input, binary.LittleEndian, uint32(len(original.Rows))); err != nil {
		t.Fatal(err)
	}
	for _, row := range original.Rows {
		var events [12][5]int64
		if len(row.Events) > len(events) {
			t.Fatal("original trace exceeds fixture")
		}
		copy(events[:], row.Events)
		for _, value := range []any{row.Input, uint32(len(row.Events)), events, row.After} {
			if err := binary.Write(&input, binary.LittleEndian, value); err != nil {
				t.Fatal(err)
			}
		}
	}
	source, err := os.ReadFile("client__gui__guigen.c")
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("testdata/client_mapgen_text_test.c")
	if err != nil {
		t.Fatal(err)
	}
	fixture = []byte(strings.Replace(string(fixture), "// MAPGEN_BODY", string(source), 1))
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "client-mapgen-text")
	args := append(compiler[1:], "-std=c11", "-O2", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing", "-fno-strict-overflow", "-Wall", "-Wextra", "-Werror", "-I.", "-Itestdata", "-x", "c", "-", "-o", bin)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = bytes.NewReader(fixture)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile complete map-generation body: %v\n%s", err, output)
	}
	cmd = exec.Command(bin)
	cmd.Stdin = bytes.NewReader(input.Bytes())
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("actual native map-generation body: %v\n%s", err, output)
	} else {
		t.Log(strings.TrimSpace(string(output)))
	}
}
