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

// Compile the complete save boundary with real native structure
// declarations. Controlled services observe pointer identity, exact object
// bytes, call order and original failure branches; no original binary runs.
func TestMapgenSaveMoveNativeR311(t *testing.T) {
	t.Run("save", func(t *testing.T) {
		const fixturePath = "testdata/mapgen_save_native_r311_test.c"
		const receipt = "Mapgen save native R311: 48 original-domain cases PASS"
		fixture, err := os.ReadFile(fixturePath)
		if err != nil {
			t.Fatal(err)
		}
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
				binary := filepath.Join(t.TempDir(), "save")
				if runtime.GOOS == "windows" {
					binary += ".exe"
				}
				args := append([]string(nil), compiler[1:]...)
				args = append(args, "-std=c11", "-g", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing", "-I.")
				args = append(args, mode.flags...)
				args = append(args, "-x", "c", "-", "-o", binary)
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				cmd := exec.CommandContext(ctx, compiler[0], args...)
				cmd.Stdin = strings.NewReader(string(fixture))
				out, err := cmd.CombinedOutput()
				cancel()
				if err != nil {
					t.Fatalf("compile complete native boundary: %v\n%s", err, out)
				}
				ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				cmd = exec.CommandContext(ctx, binary)
				out, err = cmd.CombinedOutput()
				if err != nil || !strings.Contains(string(out), receipt) {
					t.Fatalf("native contract: %v\n%s", err, out)
				}
				t.Log(strings.TrimSpace(string(out)))
			})
		}
	})
	t.Run("move", testMapgenMoveNativeR311)
}
