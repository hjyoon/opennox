package legacy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Compile the production C bodies against the real native Player/Window
// headers. The public old Host Game PNG remains a separate frozen gate;
// matching that PNG must not restore a PE32 status read on a native Player.
func TestClientObserverWindow4357A0NativeBodies(t *testing.T) {
	fixture, err := os.ReadFile("testdata/client_observer_window_4357a0_test.c")
	if err != nil {
		t.Fatal(err)
	}
	program := string(fixture)
	for _, part := range []struct {
		file, signature, marker string
	}{
		{"GAME1_2.c", "int nox_xxx_cliToggleObsWindow_4357A0() {", "// PRODUCTION_OBSERVER_TOGGLE"},
		{"GAME2_3.c", "int nox_xxx_showObserverWindow_48CA70(int a1) {", "// PRODUCTION_OBSERVER_VISIBILITY"},
	} {
		source, err := os.ReadFile(part.file)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(source), part.signature) != 1 || strings.Count(program, part.marker) != 1 {
			t.Fatalf("production signature or fixture marker is not unique: %s", part.signature)
		}
		_, body, _ := strings.Cut(string(source), part.signature)
		body, _, ok := strings.Cut(body, "\n//----- (")
		if !ok || !strings.HasSuffix(strings.TrimSpace(body), "}") {
			t.Fatalf("production function boundary missing: %s", part.signature)
		}
		program = strings.Replace(program, part.marker, part.signature+body, 1)
	}
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "client-observer-window")
	args := append(compiler[1:], "-std=c11", "-O2", "-Wall", "-Wextra", "-Werror",
		"-Itestdata", "-x", "c", "-", "-o", bin)
	cmd := exec.Command(compiler[0], nativeFixtureCompileArgs(t, args)...)
	cmd.Stdin = strings.NewReader(program)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile native observer visibility bodies: %v\n%s", err, output)
	}
	if output, err := exec.Command(bin).CombinedOutput(); err != nil {
		t.Fatalf("native observer window status contract: %v\n%s", err, output)
	} else {
		t.Log(strings.TrimSpace(string(output)))
	}
}
