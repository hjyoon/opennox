package legacy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Compile all four live state-table functions together. The reference walks
// the five original scalar records independently; only memmap relocation and
// the clock are supplied by the fixture.
func clientAbilityStateNative461090(t *testing.T, mode string) {
	t.Helper()
	source, err := os.ReadFile("GAME2_1.c")
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("testdata/client_ability_state_test.c")
	if err != nil {
		t.Fatal(err)
	}
	start, end := "//----- (00461090)", "//----- (004611A0)"
	if strings.Count(string(source), start) != 1 || strings.Count(string(source), end) != 1 {
		t.Fatal("ambiguous complete client ability-state subsystem")
	}
	_, body, _ := strings.Cut(string(source), start)
	body, _, _ = strings.Cut(body, end)
	program := strings.Replace(string(fixture), "// COMPLETE_SUBSYSTEM", "//"+body, 1)
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "client-ability-state")
	args := append(compiler[1:], "-std=c11", "-O2", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing",
		"-Wall", "-Wextra", "-Werror", "-Wno-unused-function", "-Itestdata", "-x", "c", "-", "-o", bin)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = strings.NewReader(program)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile actual client ability-state subsystem: %v\n%s", err, out)
	}
	if out, err := exec.Command(bin, mode).CombinedOutput(); err != nil {
		t.Fatalf("actual client ability %s contracts: %v\n%s", mode, err, out)
	} else {
		t.Log(strings.TrimSpace(string(out)))
	}
}

func TestClientAbility461090NativeStateAllRecords(t *testing.T) {
	clientAbilityStateNative461090(t, "state")
}

func TestClientAbility4610D0NativeResetAllRecords(t *testing.T) {
	clientAbilityStateNative461090(t, "reset")
}

func TestClientAbility461120NativeActiveAllRecords(t *testing.T) {
	clientAbilityStateNative461090(t, "active")
}

func TestClientAbility461160NativeLookupAllRecords(t *testing.T) {
	clientAbilityStateNative461090(t, "lookup")
}
