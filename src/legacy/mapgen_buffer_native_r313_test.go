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

func TestMapgenBufferNativeR313(t *testing.T) {
	raw, err := os.ReadFile("GAME4_2.c")
	if err != nil {
		t.Fatal(err)
	}
	source := strings.ReplaceAll(string(raw), "\r\n", "\n")
	const start = "int nox_xxx_mapgenAllocBuffer_5213E0() {"
	const end = "//----- (00521710)"
	if strings.Count(source, start) != 1 || strings.Count(source, end) != 1 {
		t.Fatal("complete buffer lifetime boundaries are not unique")
	}
	_, tail, _ := strings.Cut(source, start)
	body, _, _ := strings.Cut(tail, end)
	fixture, err := os.ReadFile("testdata/mapgen_buffer_native_r313_test.c")
	if err != nil {
		t.Fatal(err)
	}
	const marker = "// COMPLETE_MAPGEN_BUFFER_R313"
	if strings.Count(string(fixture), marker) != 1 {
		t.Fatal("complete buffer lifetime insertion point is not unique")
	}
	program := strings.Replace(string(fixture), marker, start+body, 1)
	for _, mode := range []struct {
		name  string
		flags []string
	}{
		{"O0", []string{"-O0"}},
		{"O2", []string{"-O2"}},
		{"sanitized", []string{"-O1", "-fsanitize=address,undefined", "-fno-sanitize-recover=all", "-fno-omit-frame-pointer"}},
	} {
		t.Run(mode.name, func(t *testing.T) {
			if mode.name == "sanitized" && runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
				t.Skip("sanitizer execution requires Darwin/Linux C runtime")
			}
			compiler := strings.Fields(os.Getenv("CC"))
			if len(compiler) == 0 {
				compiler = []string{"cc"}
			}
			binary := filepath.Join(t.TempDir(), "mapgen-buffer")
			if runtime.GOOS == "windows" {
				binary += ".exe"
			}
			args := append([]string(nil), compiler[1:]...)
			args = append(args, "-std=c11", "-g", "-Werror=int-conversion", "-I.")
			args = append(args, mode.flags...)
			args = append(args, "-x", "c", "-", "-o", binary)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			cmd := exec.CommandContext(ctx, compiler[0], args...)
			cmd.Stdin = strings.NewReader(program)
			out, err := cmd.CombinedOutput()
			cancel()
			if err != nil {
				t.Fatalf("compile complete native buffer lifetime: %v\n%s", err, out)
			}
			ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd = exec.CommandContext(ctx, binary)
			out, err = cmd.CombinedOutput()
			if err != nil || !strings.Contains(string(out), "Mapgen buffer native R313: 12 repeated lifetimes and allocation failures PASS") {
				t.Fatalf("native buffer lifetime contract: %v\n%s", err, out)
			}
			t.Log(strings.TrimSpace(string(out)))
		})
	}
}
