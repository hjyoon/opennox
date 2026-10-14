package legacy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Compile the whole live ally-health table subsystem. Only memmap relocation
// is controlled; allocation, lookup, updates and reset are actual C bodies.
func clientAllyHealthNative495020(t *testing.T, mode string) {
	t.Helper()
	source, err := os.ReadFile("GAME2_3.c")
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("testdata/client_ally_health_test.c")
	if err != nil {
		t.Fatal(err)
	}
	start, end := "//----- (00494FF0)", "//----- (00495210)"
	if strings.Count(string(source), start) != 1 || strings.Count(string(source), end) != 1 {
		t.Fatal("ambiguous complete ally-health subsystem")
	}
	_, body, _ := strings.Cut(string(source), start)
	body, _, _ = strings.Cut(body, end)
	program := strings.Replace(string(fixture), "// COMPLETE_SUBSYSTEM", "//"+body, 1)
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "client-ally-health")
	args := append(compiler[1:], "-std=c11", "-O2", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing",
		"-Wall", "-Wextra", "-Werror", "-Wno-unused-function", "-Itestdata", "-x", "c", "-", "-o", bin)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = strings.NewReader(program)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile actual ally-health subsystem: %v\n%s", err, out)
	}
	if out, err := exec.Command(bin, mode).CombinedOutput(); err != nil {
		t.Fatalf("actual ally-health %s contracts: %v\n%s", mode, err, out)
	} else {
		t.Log(strings.TrimSpace(string(out)))
	}
}

func TestClientAllyHealth494FF0NativeFreeSlotScan(t *testing.T) {
	clientAllyHealthNative495020(t, "free")
}

func TestClientAllyHealth495020NativeLookupLifecycle(t *testing.T) {
	clientAllyHealthNative495020(t, "lookup")
}

func TestClientAllyHealth495060NativeInsertionAllSlots(t *testing.T) {
	clientAllyHealthNative495020(t, "insert")
}

func TestClientAllyHealth4951C0NativeResetCompleteTable(t *testing.T) {
	clientAllyHealthNative495020(t, "reset")
}
