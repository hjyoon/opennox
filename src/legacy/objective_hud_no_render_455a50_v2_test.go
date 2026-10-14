package legacy

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Keep the complete active C functions and real headers. Only the native
// engine flag, team lookup, memmap, and GUI/display services are controlled;
// actual dedicated-server map loading is verified separately.
func testObjectiveHUDNoRender212V2(t *testing.T, mode string) {
	t.Helper()
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("testdata/objective_hud_no_render_455a50_v2.c")
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	source = strings.ReplaceAll(source, "/* CTF_BODY */", teamReportBody210(t, "GAME2.c", "char sub_455A50("))
	source = strings.ReplaceAll(source, "/* FLAGBALL_BODY */", teamReportBody210(t, "GAME2.c", "int sub_455F60("))
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	temp := t.TempDir()
	file, binary := filepath.Join(temp, "objective-hud.c"), filepath.Join(temp, "objective-hud")
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"-std=gnu11", "-O2", "-g", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing", "-Wno-incompatible-pointer-types", "-I", dir, file, "-o", binary}
	if runtime.GOARCH == "386" {
		args = append([]string{"-m32"}, args...)
	}
	if output, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile complete active bodies: %v\n%s", err, output)
	}
	if output, err := exec.Command(binary, mode).CombinedOutput(); err != nil {
		t.Fatalf("native objective HUD %s: %v\n%s", mode, err, output)
	} else {
		t.Log(strings.TrimSpace(string(output)))
	}
}

func TestCTFHUD455A50RespectsNativeNoRenderingV2(t *testing.T) {
	testObjectiveHUDNoRender212V2(t, "ctf")
}
func TestFlagBallHUD455F60RespectsNativeNoRenderingV2(t *testing.T) {
	testObjectiveHUDNoRender212V2(t, "flagball")
}
