package legacy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The complete active death-feed consumer is compiled with actual native
// headers and UTF-16 formatting. Lookup/font/Thing/image/draw services are
// controlled; no linked renderer, original executable or GUI is exercised.
func TestDeathFeedPlayerNamesNativeR284(t *testing.T) {
	source, err := os.ReadFile("GAME2_3.c")
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("testdata/death_feed_player_name_r284_test.c")
	if err != nil {
		t.Fatal(err)
	}
	text, program := string(source), string(fixture)
	start, end, marker := "int* sub_495500(int* a1) {", "//----- (004958F0)", "// PRODUCTION_DEATH_FEED"
	if strings.Count(text, start) != 1 || strings.Count(text, end) != 1 || strings.Count(program, marker) != 1 {
		t.Fatal("ambiguous complete death-feed production boundary")
	}
	_, tail, _ := strings.Cut(text, start)
	body, _, ok := strings.Cut(tail, end)
	if !ok || !strings.HasSuffix(strings.TrimSpace(body), "}") {
		t.Fatal("incomplete death-feed production body")
	}
	program = strings.Replace(program, marker, start+body, 1)
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "death-feed-player-name-r284")
	args := append(compiler[1:], "-std=c11", "-O2", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing",
		"-Wall", "-Wextra", "-Werror", "-Itestdata", "-x", "c", "-", "-o", bin)
	cmd := exec.Command(compiler[0], nativeFixtureCompileArgs(t, args)...)
	cmd.Stdin = strings.NewReader(program)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile complete death-feed production body: %v\n%s", err, out)
	}
	if out, err := exec.Command(bin).CombinedOutput(); err != nil {
		t.Fatalf("native death-feed name/font/icon contract: %v\n%s", err, out)
	} else {
		t.Log(strings.TrimSpace(string(out)))
	}
}
