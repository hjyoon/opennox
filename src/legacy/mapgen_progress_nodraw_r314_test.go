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

func TestMapgenProgressNoRenderingR314(t *testing.T) {
	source, err := os.ReadFile("GAME4_2.c")
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ReplaceAll(string(source), "\r\n", "\n")
	const start = "void nox_xxx_mapGenSetFlags_5235F0(char a1) {"
	const end = "//----- (00523670)"
	if strings.Count(text, start) != 1 || strings.Count(text, end) != 1 {
		t.Fatal("complete map-generation progress boundaries are not unique")
	}
	_, tail, _ := strings.Cut(text, start)
	body, _, ok := strings.Cut(tail, end)
	if !ok {
		t.Fatal("complete map-generation progress end is absent")
	}
	fixture, err := os.ReadFile("testdata/mapgen_progress_nodraw_r314_test.c")
	if err != nil {
		t.Fatal(err)
	}
	const marker = "// COMPLETE_MAPGEN_PROGRESS_R314"
	if strings.Count(string(fixture), marker) != 1 {
		t.Fatal("complete progress insertion marker is not unique")
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
				t.Skip("sanitizer execution requires the supported Darwin/Linux C runtime")
			}
			compiler := strings.Fields(os.Getenv("CC"))
			if len(compiler) == 0 {
				compiler = []string{"cc"}
			}
			binary := filepath.Join(t.TempDir(), "mapgen-progress")
			if runtime.GOOS == "windows" {
				binary += ".exe"
			}
			args := append([]string(nil), compiler[1:]...)
			args = append(args, "-std=c11", "-g", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing", "-Wall", "-Wextra", "-Werror", "-I.")
			args = append(args, mode.flags...)
			args = append(args, "-x", "c", "-", "-o", binary)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			cmd := exec.CommandContext(ctx, compiler[0], args...)
			cmd.Stdin = strings.NewReader(program)
			output, err := cmd.CombinedOutput()
			cancel()
			if err != nil {
				t.Fatalf("compile complete mapgen progress: %v\n%s", err, output)
			}
			for _, scenario := range []string{"headless", "editor", "rendered"} {
				t.Run(scenario, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					output, err := exec.CommandContext(ctx, binary, scenario).CombinedOutput()
					if err != nil {
						t.Fatalf("complete mapgen progress %s: %v\n%s", scenario, err, output)
					}
					if !strings.Contains(string(output), "Mapgen progress "+scenario+": PASS") {
						t.Fatalf("missing complete mapgen progress receipt: %s", output)
					}
				})
			}
		})
	}
}
