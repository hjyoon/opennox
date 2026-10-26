package legacy

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Compile the complete active incoming-player body with real native headers.
// Only its external services are controlled. Actual CTF/FlagBall joins are
// verified separately with native input, real clocks, and a headless seat.
func testIncomingTeamIterator213V2(t *testing.T, mode string) {
	t.Helper()
	compiler, err := exec.LookPath("cc")
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("testdata/ctf_incoming_team_iterator_4ddf60.c")
	if err != nil {
		t.Fatal(err)
	}
	source := strings.ReplaceAll(string(fixture), "/* COMPLETE_INCOMING_BODY */",
		teamReportBody210(t, "GAME3_2.c", "int nox_xxx_netPlayerIncomingServ_4DDF60("))
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	temp := t.TempDir()
	file, binary := filepath.Join(temp, "incoming-player.c"), filepath.Join(temp, "incoming-player")
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"-std=gnu11", "-O2", "-g", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing", "-fno-strict-overflow", "-Wno-incompatible-pointer-types", "-Wno-int-conversion", "-I", dir, file, "-o", binary}
	if runtime.GOARCH == "386" {
		args = append([]string{"-m32"}, args...)
	}
	if output, err := exec.Command(compiler, nativeFixtureCompileArgs(t, args)...).CombinedOutput(); err != nil {
		t.Fatalf("compile complete incoming body: %v\n%s", err, output)
	}
	if output, err := exec.Command(binary, mode).CombinedOutput(); err != nil {
		t.Fatalf("native incoming body %s: %v\n%s", mode, err, output)
	} else {
		t.Log(strings.TrimSpace(string(output)))
	}
}

func TestCTFIncoming4DDF60PreservesNativeTeamIteratorV2(t *testing.T) {
	testIncomingTeamIterator213V2(t, "ctf")
}

func TestIncoming4DDF60KeepsFlagBallPriorityAndOrdinaryTailV2(t *testing.T) {
	testIncomingTeamIterator213V2(t, "unchanged")
}
