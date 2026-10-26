package legacy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Compile the complete live selector, marker reset, code encoder and sender.
// Only grid placement and the final network sink are fixture services.
func clientSecondaryGridNative4619F0(t *testing.T, mode string) {
	t.Helper()
	program, err := os.ReadFile("testdata/client_secondary_grid_test.c")
	if err != nil {
		t.Fatal(err)
	}
	for _, part := range []struct {
		file, start, end, marker string
	}{
		{"GAME2_1.c", "//----- (00461550)", "//----- (004615C0)", "// SELECT_BODY"},
		{"GAME2_1.c", "//----- (004619F0)", "//----- (00461B50)", "// CLEAR_BODY"},
		{"GAME3_1.c", "//----- (004BF010)", "int nox_client_getPlayerDollState_4BF7E0", "// REPORT_BODY"},
		{"GAME5_2.c", "//----- (00578B00)", "//----- (00578B30)", "// CODE_BODY"},
	} {
		source, err := os.ReadFile(part.file)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(source), part.start) != 1 || strings.Count(string(source), part.end) != 1 {
			t.Fatalf("ambiguous complete secondary-weapon body in %s", part.file)
		}
		_, body, _ := strings.Cut(string(source), part.start)
		body, _, _ = strings.Cut(body, part.end)
		program = []byte(strings.Replace(string(program), part.marker, "//"+body, 1))
	}
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "client-secondary-grid")
	args := append(compiler[1:], "-std=c11", "-O2", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing",
		"-Wall", "-Wextra", "-Werror", "-Wno-unused-function", "-Itestdata", "-x", "c", "-", "-o", bin)
	cmd := exec.Command(compiler[0], nativeFixtureCompileArgs(t, args)...)
	cmd.Stdin = strings.NewReader(string(program))
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile actual secondary-weapon subsystem: %v\n%s", err, out)
	}
	if out, err := exec.Command(bin, mode).CombinedOutput(); err != nil {
		t.Fatalf("actual secondary-weapon %s contracts: %v\n%s", mode, err, out)
	} else {
		t.Log(strings.TrimSpace(string(out)))
	}
}

func TestClientSecondary4619F0NativeAll84Markers(t *testing.T) {
	clientSecondaryGridNative4619F0(t, "clear")
}

func TestClientSecondary461550NativeSelectionAndWire(t *testing.T) {
	clientSecondaryGridNative4619F0(t, "select")
}
