package legacy

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unsafe"
)

// The original v4 bound searches compare the entire first BYTE to zero, while
// the v3 whole-map input accepts that BYTE without masking. Compile both actual
// codecs together so the high flag state reaches the v4 writer through v3 I/O.
// Original GAME.EXE: 0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a;
// predicate sites 0x4222b5, 0x422305, 0x422333, 0x422370; containing candidate
// bytes [0x422230,0x422a35): 6186dbc0a9b10d1bba9fdf9cf8d7a8ca25a0e051520ad2a7c2cd6ad2ef895da4.
func TestFloorMapV4WholeFlagBoundsR303(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("64-bit native pointer regression")
	}
	fixture, err := os.ReadFile("testdata/floor_map_whole_flag_bounds_r303_test.c")
	if err != nil {
		t.Fatal(err)
	}
	var bodies strings.Builder
	for _, part := range []struct{ file, symbol string }{
		{"GAME1_1.c", "nox_server_mapRWFloorMap_422230"},
		{"GAME1_1.c", "nox_xxx_tileReadOne_422A40"},
		{"GAME1_1.c", "nox_xxx_tile_422C10"},
		{"GAME1_2.c", "sub_428170"},
	} {
		data, err := os.ReadFile(part.file)
		if err != nil {
			t.Fatal(err)
		}
		starts := regexp.MustCompile(`(?m)^(?:int|unsigned char) `+part.symbol+`\([^;\n]*\) \{`).FindAllIndex(data, -1)
		if len(starts) != 1 {
			t.Fatalf("nonunique production function %s", part.symbol)
		}
		body, _, ok := strings.Cut(string(data[starts[0][0]:]), "\n}\n")
		if !ok {
			t.Fatalf("missing production end for %s", part.symbol)
		}
		bodies.WriteString(body + "\n}\n")
	}
	program := strings.Replace(string(fixture), "// PRODUCTION_FLOOR_WHOLE_FLAG_BODIES", bodies.String(), 1)
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "floor-whole-flag-bounds")
	args := append(compiler[1:], "-std=c11", "-O2", "-Werror", "-Wno-incompatible-pointer-types", "-I.", "-x", "c", "-", "-o", bin)
	cmd := exec.Command(compiler[0], nativeFixtureCompileArgs(t, args)...)
	cmd.Stdin = strings.NewReader(program)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile actual FloorMap input and writer: %v\n%s", err, out)
	}
	for _, name := range []string{"high-only-04", "high-only-80", "high-only-grid-edge", "high-and-layer-05", "high-and-two-layers-07", "layer-clear-high-only", "low-layer-control", "empty-control", "adjacent-byte-0100", "adjacent-byte-80000000"} {
		t.Run(name, func(t *testing.T) {
			if out, err := exec.Command(bin, name).CombinedOutput(); err != nil {
				t.Fatalf("v3 input to v4 whole-BYTE bounds: %v\n%s", err, out)
			} else {
				t.Log(strings.TrimSpace(string(out)))
			}
		})
	}
}
