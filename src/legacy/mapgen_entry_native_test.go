package legacy

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestMapgenEntryNativeObjectList(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("this contract requires native allocations above the PE32 address range")
	}
	fixture, err := os.ReadFile("testdata/mapgen_entry_native_test.c")
	if err != nil {
		t.Fatal(err)
	}
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
			binary := filepath.Join(t.TempDir(), "mapgen-entry")
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
			output, err := cmd.CombinedOutput()
			cancel()
			if err != nil {
				t.Fatalf("compile complete generator entry: %v\n%s", err, output)
			}
			for _, name := range []string{"empty", "generated"} {
				t.Run(name, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					cmd := exec.CommandContext(ctx, binary, name)
					output, err := cmd.CombinedOutput()
					if err != nil {
						t.Fatalf("complete generator native pointer contract: %v\n%s", err, output)
					}
					if !strings.Contains(string(output), "Classic entry native pointers and script_id: "+name+" PASS") {
						t.Fatalf("missing complete generator entry receipt: %s", output)
					}
				})
			}
		})
	}
}
