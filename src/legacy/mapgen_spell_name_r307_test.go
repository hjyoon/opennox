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

// Execute the complete production body and the actual noxstring formatter.
// O0/O2 observe the callee object's capacity before a potentially invalid
// write. ASan+UBSan execute the actual formatting write without that guard.
// All modes preserve the same 310 bounded inputs, expected text and lookup
// contracts. This is not an execution of the complete classic generator.
func TestMapgenSpellNameOriginalBuffersR307(t *testing.T) {
	source, err := os.ReadFile("GAME4_2.c")
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("testdata/mapgen_spell_name_r307_test.c")
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ReplaceAll(string(source), "\r\n", "\n")
	start := "int nox_xxx_mapGenSpellIdByName_51E1D0("
	end := "\n// 51E1D0:"
	if strings.Count(text, start) != 1 || strings.Count(text, end) != 1 {
		t.Fatal("nonunique complete production spell-name body")
	}
	_, tail, _ := strings.Cut(text, start)
	body, _, ok := strings.Cut(tail, end)
	if !ok {
		t.Fatal("missing production spell-name body boundary")
	}
	const marker = "// PRODUCTION_SPELL_NAME_51E1D0"
	program := strings.ReplaceAll(string(fixture), "\r\n", "\n")
	if strings.Count(program, marker) != 1 {
		t.Fatal("nonunique production spell-name fixture marker")
	}
	program = strings.Replace(program, marker, start+body, 1)
	modes := []struct {
		name, contract string
		flags          []string
	}{
		{"O0", "capacity", []string{"-O0"}},
		{"O2", "capacity", []string{"-O2"}},
		{"sanitized", "native", []string{"-O1", "-fsanitize=address,undefined", "-fno-sanitize-recover=all", "-fno-omit-frame-pointer"}},
	}
	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			if mode.name == "sanitized" && runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
				t.Skip("sanitizer mode requires the supported Darwin/Linux C runtime")
			}
			compiler := strings.Fields(os.Getenv("CC"))
			if len(compiler) == 0 {
				compiler = []string{"cc"}
			}
			bin := filepath.Join(t.TempDir(), "mapgen-spell-name")
			if runtime.GOOS == "windows" {
				bin += ".exe"
			}
			args := append([]string(nil), compiler[1:]...)
			args = append(args, "-std=c11", "-g", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing",
				"-Wall", "-Wextra", "-Werror", "-Wno-unused-function")
			if runtime.GOOS == "darwin" {
				args = append(args, "-Wno-deprecated-declarations")
			}
			args = append(args, mode.flags...)
			args = append(args, "-I.", "-x", "c", "-", "-o", bin)
			env := make([]string, 0, len(os.Environ())+2)
			for _, value := range os.Environ() {
				if !strings.HasPrefix(value, "ASAN_OPTIONS=") && !strings.HasPrefix(value, "UBSAN_OPTIONS=") {
					env = append(env, value)
				}
			}
			env = append(env, "ASAN_OPTIONS=detect_leaks=0:halt_on_error=1:abort_on_error=0",
				"UBSAN_OPTIONS=halt_on_error=1:print_stacktrace=1")
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			cmd := exec.CommandContext(ctx, compiler[0], args...)
			cmd.Stdin = strings.NewReader(program)
			cmd.Env = env
			out, err := cmd.CombinedOutput()
			cancel()
			if err != nil {
				t.Fatalf("compile complete production spell-name body: %v\n%s", err, out)
			}
			ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd = exec.CommandContext(ctx, bin, mode.contract)
			cmd.Env = env
			out, err = cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("production spell-name buffer/format/lookup contract: %v\n%s", err, out)
			}
			if !strings.Contains(string(out), "Spell-name 51E1D0: 310 bounded cases;") {
				t.Fatalf("missing complete spell-name fixture result: %s", out)
			}
			t.Log(strings.TrimSpace(string(out)))
		})
	}
}
