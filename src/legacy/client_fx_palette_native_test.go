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

func clientFXPaletteNative4B5CD0(t *testing.T, mode string) {
	t.Helper()
	var oracle struct {
		OriginalSHA string      `json:"original_sha256"`
		CodeBytes   int         `json:"code_bytes"`
		CodeSHA     string      `json:"code_sha256"`
		Calls       [][4]uint32 `json:"calls"`
		Words       []uint32    `json:"table_words"`
		Returned    uint32      `json:"returned"`
	}
	raw, err := os.ReadFile("testdata/fx_palette_4b5cd0_original.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &oracle); err != nil {
		t.Fatal(err)
	}
	if oracle.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || oracle.CodeBytes != 319 || oracle.CodeSHA != "b193fadf653a19c714b42b2a6cfa0ca883e967e9946ad32adde4a621659f39dc" || len(oracle.Calls) != 256 || len(oracle.Words) != 256 {
		t.Fatal("wrong sealed original palette execution")
	}
	var input bytes.Buffer
	for _, values := range []any{oracle.Calls, oracle.Words, oracle.Returned} {
		if err := binary.Write(&input, binary.LittleEndian, values); err != nil {
			t.Fatal(err)
		}
	}
	program, err := os.ReadFile("testdata/client_fx_palette_test.c")
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("GAME3.c")
	if err != nil {
		t.Fatal(err)
	}
	const start, end = "//----- (004B5CD0)", "//----- (004B63B0)"
	if strings.Count(string(source), start) != 1 || strings.Count(string(source), end) != 1 {
		t.Fatal("ambiguous complete palette body")
	}
	_, body, _ := strings.Cut(string(source), start)
	body, _, _ = strings.Cut(body, end)
	program = []byte(strings.Replace(string(program), "// PALETTE_BODY", "//"+body, 1))
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "client-fx-palette")
	args := append(compiler[1:], "-std=c11", "-O2", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing", "-Wall", "-Wextra", "-Werror", "-Itestdata", "-x", "c", "-", "-o", bin)
	cmd := exec.Command(compiler[0], nativeFixtureCompileArgs(t, args)...)
	cmd.Stdin = bytes.NewReader(program)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile actual palette body: %v\n%s", err, output)
	}
	cmd = exec.Command(bin, mode)
	cmd.Stdin = bytes.NewReader(input.Bytes())
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("actual native palette %s: %v\n%s", mode, err, output)
	} else {
		t.Log(strings.TrimSpace(string(output)))
	}
}

func TestClientFXPalette4B5CD0OriginalRGBBytes(t *testing.T) {
	clientFXPaletteNative4B5CD0(t, "colors")
}

func TestClientFXPalette4B5CD0NativeFourLoopBounds(t *testing.T) {
	clientFXPaletteNative4B5CD0(t, "bounds")
}
