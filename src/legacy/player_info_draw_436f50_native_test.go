package legacy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Execute the complete production drawer with native layouts and the real
// UTF-16 formatter. The frozen PE32 table deliberately retains the exact
// adjacent addresses from the reported Linux/amd64 crash.
func TestPlayerInfoDraw436F50NativePointers(t *testing.T) {
	source, err := os.ReadFile("client__system__client.c")
	if err != nil {
		t.Fatal(err)
	}
	const start, end = "//----- (00436F50)", "//----- (004358A0)"
	if strings.Count(string(source), start) != 1 || strings.Count(string(source), end) != 1 {
		t.Fatal("missing or ambiguous complete player info drawer")
	}
	_, body, _ := strings.Cut(string(source), start)
	body, _, _ = strings.Cut(body, end)
	fixture, err := os.ReadFile("testdata/player_info_draw_436f50_native_test.c")
	if err != nil {
		t.Fatal(err)
	}
	const insertion = "// COMPLETE_PLAYER_INFO_DRAWER"
	if strings.Count(string(fixture), insertion) != 1 {
		t.Fatal("ambiguous native fixture insertion")
	}
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "player-info-draw")
	args := append(compiler[1:], "-std=c11", "-O2", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing",
		"-Wall", "-Wextra", "-Werror", "-Wno-unused-function", "-Itestdata", "-x", "c", "-", "-o", bin)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = strings.NewReader(strings.Replace(string(fixture), insertion, "//"+body, 1))
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile complete player info drawer: %v\n%s", err, out)
	}
	for _, name := range []string{"WizardReportedPointer", "Warrior", "Conjurer", "NoLocalPlayer", "NoPlayerInfo"} {
		t.Run(name, func(t *testing.T) {
			if out, err := exec.Command(bin, name).CombinedOutput(); err != nil {
				t.Fatalf("native player info %s: %v\n%s", name, err, out)
			} else {
				t.Log(strings.TrimSpace(string(out)))
			}
		})
	}
}
