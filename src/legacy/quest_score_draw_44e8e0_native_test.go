package legacy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Compile the complete live C renderer with real native layouts and UTF-16
// formatting. Only the memory map, locale and drawing services are controlled.
func questScoreDrawNative44E8E0(t *testing.T, mode string) {
	t.Helper()
	source, err := os.ReadFile("client__gui__guibrief.c")
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("testdata/quest_score_draw_44e8e0_test.c")
	if err != nil {
		t.Fatal(err)
	}
	start, end := "//----- (0044E8E0)", "//----- (0044F0F0)"
	if strings.Count(string(source), start) != 1 || strings.Count(string(source), end) != 1 {
		t.Fatal("ambiguous complete Quest score renderer")
	}
	_, body, _ := strings.Cut(string(source), start)
	body, _, _ = strings.Cut(body, end)
	program := strings.Replace(string(fixture), "// COMPLETE_RENDERER", "//"+body, 1)
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "quest-score-draw")
	args := append(compiler[1:], "-std=c11", "-O2", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing",
		"-Wall", "-Wextra", "-Werror", "-Wno-unused-function", "-Itestdata", "-x", "c", "-", "-o", bin)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = strings.NewReader(program)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile complete Quest score renderer: %v\n%s", err, out)
	}
	if out, err := exec.Command(bin, mode).CombinedOutput(); err != nil {
		t.Fatalf("complete Quest score %s contracts: %v\n%s", mode, err, out)
	} else {
		t.Log(strings.TrimSpace(string(out)))
	}
}

func TestQuestScoreDraw44E8E0OriginalStageSeparator(t *testing.T) {
	questScoreDrawNative44E8E0(t, "title")
}

func TestQuestScoreDraw44E8E0NativeSixRowBoundary(t *testing.T) {
	questScoreDrawNative44E8E0(t, "rows")
}
