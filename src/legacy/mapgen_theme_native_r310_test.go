package legacy

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// The complete producers execute with the actual token registry and controlled
// file/token/formatter/lookup/settings callbacks. Fixed expectations cover the
// original offsets, dispatch, partial writes and original bounded input domain.
// A real encrypted-binfile/CGo integration is exercised separately.
func TestMapgenThemeNativeOriginalContractsR310(t *testing.T) {
	source, err := os.ReadFile("GAME4_2.c")
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ReplaceAll(string(source), "\r\n", "\n")
	var bodies []string
	for _, item := range []struct{ start, end string }{
		{"int nox_xxx_mapGenSpellIdByName_51E1D0(", "\n// 51E1D0:"},
		{"int nox_mapgenReadAlgDataNative_51EBB0(", "\n//----- (0051EFB0)"},
		{"int nox_mapgenReadSpellSetNative_51EFB0(", "\n//----- (0051F030)"},
		{"int nox_mapgenReadExitNative_51F800(", "\n//----- (0051F9F0)"},
		{"int nox_mapgenReadThemeNative_51E260(", "\n//----- (0051E540)"},
	} {
		if strings.Count(text, item.start) != 1 || strings.Count(text, item.end) != 1 {
			t.Fatalf("nonunique complete production body: %s", item.start)
		}
		_, tail, _ := strings.Cut(text, item.start)
		body, _, ok := strings.Cut(tail, item.end)
		if !ok {
			t.Fatalf("missing production body boundary: %s", item.start)
		}
		bodies = append(bodies, item.start+body)
	}
	fixture, err := os.ReadFile("testdata/mapgen_theme_native_r310_test.c")
	if err != nil {
		t.Fatal(err)
	}
	program := strings.ReplaceAll(string(fixture), "\r\n", "\n")
	const marker = "// PRODUCTION_THEME_PARSERS_R310"
	if strings.Count(program, marker) != 1 {
		t.Fatal("nonunique complete producer insertion point")
	}
	program = strings.Replace(program, marker, strings.Join(bodies, "\n"), 1)
	for _, mode := range []struct {
		name, contract string
		flags          []string
	}{
		{"O0", "capacity", []string{"-O0"}},
		{"O2", "capacity", []string{"-O2"}},
		{"sanitized", "native", []string{"-O1", "-fsanitize=address,undefined", "-fno-sanitize-recover=all", "-fno-omit-frame-pointer"}},
	} {
		t.Run(mode.name, func(t *testing.T) {
			if mode.name == "sanitized" && runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
				t.Skip("sanitizer execution requires supported Darwin/Linux C runtime")
			}
			compiler := strings.Fields(os.Getenv("CC"))
			if len(compiler) == 0 {
				compiler = []string{"cc"}
			}
			binary := filepath.Join(t.TempDir(), "theme-native")
			if runtime.GOOS == "windows" {
				binary += ".exe"
			}
			args := append([]string(nil), compiler[1:]...)
			args = append(args, "-std=c11", "-D_POSIX_C_SOURCE=200809L", "-g", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing",
				"-Wall", "-Wextra", "-Werror", "-Wno-unused-function", "-Wno-unused-parameter", "-Wno-logical-op-parentheses")
			if runtime.GOOS == "darwin" {
				args = append(args, "-Wno-deprecated-declarations")
			}
			args = append(args, mode.flags...)
			args = append(args, "-I.", "-x", "c", "-", "-o", binary)
			var env []string
			for _, value := range os.Environ() {
				if !strings.HasPrefix(value, "ASAN_OPTIONS=") && !strings.HasPrefix(value, "UBSAN_OPTIONS=") {
					env = append(env, value)
				}
			}
			env = append(env, "ASAN_OPTIONS=detect_leaks=0:halt_on_error=1:abort_on_error=0", "UBSAN_OPTIONS=halt_on_error=1:print_stacktrace=1")
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			cmd := exec.CommandContext(ctx, compiler[0], args...)
			cmd.Stdin, cmd.Env = strings.NewReader(program), env
			out, err := cmd.CombinedOutput()
			cancel()
			if err != nil {
				t.Fatalf("compile complete theme producers: %v\n%s", err, out)
			}
			ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd = exec.CommandContext(ctx, binary, mode.contract)
			cmd.Env = env
			out, err = cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("native theme producer contract: %v\n%s", err, out)
			}
			if !strings.Contains(string(out), "Theme native R310: 236 bounded cases;") {
				t.Fatalf("missing complete fixture result: %s", out)
			}
			t.Log(strings.TrimSpace(string(out)))
		})
	}
}
