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
	source, err := os.ReadFile("GAME3_2.c")
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ReplaceAll(string(source), "\r\n", "\n")
	const start = "int nox_xxx_mapGenStart_4D4320() {"
	const end = "//----- (004D44E0)"
	if strings.Count(text, start) != 1 || strings.Count(text, end) != 1 {
		t.Fatal("complete generator entry boundaries are not unique")
	}
	_, tail, _ := strings.Cut(text, start)
	body, _, ok := strings.Cut(tail, end)
	if !ok {
		t.Fatal("complete generator entry end is absent")
	}
	fixture, err := os.ReadFile("testdata/mapgen_entry_native_test.c")
	if err != nil {
		t.Fatal(err)
	}
	const marker = "// COMPLETE_MAPGEN_ENTRY"
	if strings.Count(string(fixture), marker) != 1 {
		t.Fatal("complete generator insertion marker is not unique")
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
			cmd.Stdin = strings.NewReader(program)
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
