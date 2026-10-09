package legacy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestTrapQuickbarVisibility460B90NativeBodies(t *testing.T) {
	fixture, err := os.ReadFile("testdata/trap_quickbar_visibility_460b90_test.c")
	if err != nil {
		t.Fatal(err)
	}
	program := string(fixture)
	for _, part := range []struct {
		file, signature, end, marker string
	}{
		{"GAME2.c", "int sub_460B90(int a1) {", "\n}", "// PRODUCTION_VISIBILITY"},
		{"GAME2_1.c", "void sub_461010() {", "\n//----- (00461060)", "// PRODUCTION_CLOSE"},
		{"GAME2_1.c", "void sub_461060() {", "\n//----- (00461090)", "// PRODUCTION_OPEN"},
	} {
		source, err := os.ReadFile(part.file)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(source), part.signature) != 1 || strings.Count(program, part.marker) != 1 {
			t.Fatalf("production signature or marker is not unique: %q", part.signature)
		}
		_, body, _ := strings.Cut(string(source), part.signature)
		body, _, ok := strings.Cut(body, part.end)
		if !ok {
			t.Fatalf("production body boundary is missing: %q", part.signature)
		}
		if part.end == "\n}" {
			body += "\n}"
		}
		if !strings.HasSuffix(strings.TrimSpace(body), "}") {
			t.Fatalf("production body is not complete: %q", part.signature)
		}
		program = strings.Replace(program, part.marker, part.signature+body, 1)
	}
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	binary := filepath.Join(t.TempDir(), "trap-quickbar-visibility")
	args := append(compiler[1:], "-std=c11", "-O2", "-Wall", "-Wextra", "-Werror",
		"-Itestdata", "-x", "c", "-", "-o", binary)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = strings.NewReader(program)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile native Trap Set bodies: %v\n%s", err, output)
	}
	if output, err := exec.Command(binary).CombinedOutput(); err != nil {
		t.Fatalf("native Trap Set visibility contract: %v\n%s", err, output)
	} else {
		t.Log(strings.TrimSpace(string(output)))
	}
}
