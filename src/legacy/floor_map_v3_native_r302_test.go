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

// Exercise the complete production v3 codec with nonempty native tile rows and
// lists. Controlled services make the byte stream and callback order observable.
func TestFloorMapV3NativeDataR302(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("64-bit native pointer regression")
	}
	fixture, err := os.ReadFile("testdata/floor_map_v3_native_r302_test.c")
	if err != nil {
		t.Fatal(err)
	}
	var bodies strings.Builder
	for _, part := range []struct{ file, symbol string }{
		{"GAME1_1.c", "nox_xxx_tile_422C10"},
		{"GAME1_2.c", "sub_428170"},
	} {
		data, err := os.ReadFile(part.file)
		if err != nil {
			t.Fatal(err)
		}
		start := regexp.MustCompile(`(?m)^int `+part.symbol+`\([^;\n]*\) \{`).FindAllIndex(data, -1)
		if len(start) != 1 {
			t.Fatalf("nonunique production function %s", part.symbol)
		}
		body, _, ok := strings.Cut(string(data[start[0][0]:]), "\n}\n")
		if !ok {
			t.Fatalf("missing production end for %s", part.symbol)
		}
		bodies.WriteString(body + "\n}\n")
	}
	program := strings.Replace(string(fixture), "// PRODUCTION_FLOOR_V3_BODIES", bodies.String(), 1)
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "floor-map-v3-native")
	args := append(compiler[1:], "-std=c11", "-O2", "-Werror", "-Wno-incompatible-pointer-types",
		"-I.", "-x", "c", "-", "-o", bin)
	cmd := exec.Command(compiler[0], nativeFixtureCompileArgs(t, args)...)
	cmd.Stdin = strings.NewReader(program)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile actual v3 FloorMap function: %v\n%s", err, out)
	}
	for _, name := range []string{"nil-region-write", "geometry-region-write", "nil-region-read", "region-read", "edge-write", "edge-read", "empty-write"} {
		t.Run(name, func(t *testing.T) {
			if out, err := exec.Command(bin, name).CombinedOutput(); err != nil {
				t.Fatalf("native v3 FloorMap contract: %v\n%s", err, out)
			} else {
				t.Log(strings.TrimSpace(string(out)))
			}
		})
	}
}
