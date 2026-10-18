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

// Compile each complete production boundary with real native structure
// declarations. Controlled services observe pointer identity, exact object
// bytes, call order and original failure branches; no original binary runs.
func TestMapgenSaveMoveNativeR311(t *testing.T) {
	raw, err := os.ReadFile("GAME4_2.c")
	if err != nil {
		t.Fatal(err)
	}
	source := strings.ReplaceAll(string(raw), "\r\n", "\n")
	for _, item := range []struct{ name, start, end, fixture, marker, receipt string }{
		{"save", "int nox_xxx_mapSaveMap_51E010(", "\n// 51E0D5:", "testdata/mapgen_save_native_r311_test.c", "// COMPLETE_MAPGEN_SAVE_R311", "Mapgen save native R311: 48 original-domain cases PASS"},
		{"move", "float* nox_xxx_mapGenMoveObject_527A10(", "\n//----- (00527C60)", "testdata/mapgen_move_native_r311_test.c", "// COMPLETE_MAPGEN_MOVE_R311", "Mapgen move native R311: 18 original-domain cases PASS"},
	} {
		t.Run(item.name, func(t *testing.T) {
			if strings.Count(source, item.start) != 1 || strings.Count(source, item.end) != 1 {
				t.Fatal("complete production boundaries are not unique")
			}
			_, tail, _ := strings.Cut(source, item.start)
			body, _, _ := strings.Cut(tail, item.end)
			fixture, err := os.ReadFile(item.fixture)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(fixture), item.marker) != 1 {
				t.Fatal("fixture insertion point is not unique")
			}
			program := strings.Replace(string(fixture), item.marker, item.start+body, 1)
			for _, mode := range []struct {
				name  string
				flags []string
			}{
				{"O0", []string{"-O0"}}, {"O2", []string{"-O2"}},
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
					binary := filepath.Join(t.TempDir(), item.name)
					if runtime.GOOS == "windows" {
						binary += ".exe"
					}
					args := append([]string(nil), compiler[1:]...)
					args = append(args, "-std=c11", "-g", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing", "-I.")
					args = append(args, mode.flags...)
					args = append(args, "-x", "c", "-", "-o", binary)
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					cmd := exec.CommandContext(ctx, compiler[0], args...)
					cmd.Stdin = strings.NewReader(program)
					out, err := cmd.CombinedOutput()
					cancel()
					if err != nil {
						t.Fatalf("compile complete native boundary: %v\n%s", err, out)
					}
					ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					cmd = exec.CommandContext(ctx, binary)
					out, err = cmd.CombinedOutput()
					if err != nil || !strings.Contains(string(out), item.receipt) {
						t.Fatalf("native contract: %v\n%s", err, out)
					}
					t.Log(strings.TrimSpace(string(out)))
				})
			}
		})
	}
}
