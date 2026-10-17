package legacy

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
	"unsafe"
)

// These are source-bound native C callback contracts, not a complete classic
// mapgen run or a Go/CGo boundary test. Original GAME.EXE SHA256:
// 0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a.
// REA/Ghidra complete bodies: 524950..5249B7, 5249C0..524B40,
// 524B50..524DF6. FMUL/FADD operands are binary32 583B9C=41821B8F,
// 583B94=42021B8F, and 583D18=42821B8F. The literal vectors preserve
// callback order, x87-to-float staging, signed division, and RNG ranges.
// The retained int return checks preserve the current C return expression;
// Ghidra's original void inference does not establish an original int result.
func floorFillBodyR305(t *testing.T, source []byte, signature, name string) string {
	t.Helper()
	starts := regexp.MustCompile(signature).FindAllIndex(source, -1)
	if len(starts) != 1 {
		t.Fatalf("nonunique production definition %s", name)
	}
	body, _, ok := strings.Cut(string(source[starts[0][0]:]), "\n}\n")
	if !ok {
		t.Fatalf("missing complete production body %s", name)
	}
	return body + "\n}\n"
}

func floorFillProgramR305(t *testing.T) (string, []byte, []byte) {
	t.Helper()
	fixture, err := os.ReadFile("testdata/floor_fill_points_r305_test.c")
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("GAME4_2.c")
	if err != nil {
		t.Fatal(err)
	}
	typeStarts := regexp.MustCompile(`(?m)^typedef struct nox_mapgen_decor_fill_legacy \{`).FindAllIndex(source, -1)
	if len(typeStarts) != 1 {
		t.Fatal("nonunique production fill record")
	}
	record, _, ok := strings.Cut(string(source[typeStarts[0][0]:]), "\n} nox_mapgen_decor_fill_legacy;")
	if !ok {
		t.Fatal("missing complete production fill record")
	}
	types := record + "\n} nox_mapgen_decor_fill_legacy;\n" + floorFillBodyR305(t, source,
		`(?m)^static nox_mapgen_decor_fill_legacy\* nox_mapgenDecorFillResolve\([^;{}]*\)\s*\{`, "nox_mapgenDecorFillResolve")
	var bodies strings.Builder
	for _, name := range []string{"sub_524950", "sub_5249C0", "sub_524B50"} {
		bodies.WriteString(floorFillBodyR305(t, source,
			`(?m)^(?:int|void)\s+`+regexp.QuoteMeta(name)+`\([^;{}]*\)\s*\{`, name))
	}
	program := string(fixture)
	for marker, replacement := range map[string]string{
		"// PRODUCTION_FILL_POINT_TYPES_R305":  types,
		"// PRODUCTION_FILL_POINT_BODIES_R305": bodies.String(),
	} {
		if strings.Count(program, marker) != 1 {
			t.Fatalf("nonunique production insertion marker %s", marker)
		}
		program = strings.Replace(program, marker, replacement, 1)
	}
	registry, err := os.ReadFile("mapgen_legacy_ptr.c")
	if err != nil {
		t.Fatal(err)
	}
	header, err := os.ReadFile("mapgen_legacy_ptr.h")
	if err != nil {
		t.Fatal(err)
	}
	return program, registry, header
}

func floorFillEnvironmentR305() []string {
	env := make([]string, 0, len(os.Environ())+2)
	for _, value := range os.Environ() {
		if !strings.HasPrefix(value, "ASAN_OPTIONS=") && !strings.HasPrefix(value, "UBSAN_OPTIONS=") {
			env = append(env, value)
		}
	}
	// Darwin does not support detect_leaks. This regression checks address and
	// undefined behavior; it makes no leak-detection claim.
	return append(env, "ASAN_OPTIONS=detect_leaks=0:halt_on_error=1:abort_on_error=0",
		"UBSAN_OPTIONS=halt_on_error=1:print_stacktrace=1")
}

func TestFloorBrushFillPointsOriginalR305(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("64-bit native fill pointer regression")
	}
	program, registry, header := floorFillProgramR305(t)
	cases := []string{
		"sub_524950-null-fill",
		"sub_5249C0-null-fill",
		"sub_524B50-null-fill",
		"rectangle-center-negative-tie",
		"rectangle-zero-dimension-still-calls",
		"taper-odd-height-row-order",
		"taper-first-pass-extended-second-pass-staged",
		"taper-negative-one-height-truncates-to-zero",
		"taper-zero-height-includes-top-row",
		"irregular-three-square-no-rng",
		"irregular-rng-and-column-order",
		"irregular-three-by-four-mixed-branches",
		"irregular-width-two-skips-all",
		"irregular-height-two-skips-all",
		"irregular-negative-size-skips-all",
	}
	modes := []struct {
		name  string
		flags []string
	}{
		{"optimized", []string{"-O2"}},
		{"sanitized", []string{"-O1", "-fsanitize=address,undefined", "-fno-omit-frame-pointer"}},
	}
	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			if mode.name == "sanitized" && runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
				t.Skip("sanitizer mode requires the supported Darwin/Linux C runtime")
			}
			dir := t.TempDir()
			for name, data := range map[string][]byte{
				"floor-fill-points.c": []byte(program),
				"mapgen_legacy_ptr.c": registry,
				"mapgen_legacy_ptr.h": header,
			} {
				if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			compiler := strings.Fields(os.Getenv("CC"))
			if len(compiler) == 0 {
				compiler = []string{"cc"}
			}
			bin := filepath.Join(dir, "floor-fill-points")
			args := append([]string(nil), compiler[1:]...)
			args = append(args, "-std=c11", "-g", "-Wall", "-Wextra", "-Werror",
				"-Wno-unused-parameter", "-Wno-unused-but-set-variable", "-ffp-contract=off")
			args = append(args, mode.flags...)
			args = append(args, "-I"+dir, filepath.Join(dir, "floor-fill-points.c"),
				filepath.Join(dir, "mapgen_legacy_ptr.c"), "-o", bin)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			cmd := exec.CommandContext(ctx, compiler[0], args...)
			cmd.Env = floorFillEnvironmentR305()
			out, err := cmd.CombinedOutput()
			cancel()
			if err != nil {
				t.Fatalf("compile production fill points and actual registry: %v\n%s", err, out)
			}
			for _, name := range cases {
				t.Run(name, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					cmd := exec.CommandContext(ctx, bin, name)
					cmd.Env = floorFillEnvironmentR305()
					if out, err := cmd.CombinedOutput(); err != nil {
						t.Fatalf("native fill callback/token/canary contract: %v\n%s", err, out)
					} else {
						t.Log(strings.TrimSpace(string(out)))
					}
				})
			}
		})
	}
}
