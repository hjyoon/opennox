package legacy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Extract both complete production switch arms and their actual Player
// declarations. This fixture does not manufacture a corrected body; the
// unrelated dispatcher arms remain a separate actual-CGo regression.
func TestClientScorePackets48EA70NativeBodies(t *testing.T) {
	source, err := os.ReadFile("client__network__cdecode.c")
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("testdata/client_score_packet_48ea70_test.c")
	if err != nil {
		t.Fatal(err)
	}
	text, program := string(source), string(fixture)
	for _, name := range []string{"v127", "v142"} {
		matches := regexp.MustCompile(`(?m)^\t(?:char\*|nox_playerInfo\*) `+name+`;[^\n]*`).FindAllString(text, -1)
		if len(matches) != 1 {
			t.Fatalf("ambiguous Player declaration: %s", name)
		}
		program = strings.Replace(program, "// PRODUCTION_PLAYER_"+name, matches[0], 1)
	}
	for _, op := range []int{78, 85} {
		start, end := fmt.Sprintf("\tcase %d:", op), fmt.Sprintf("\tcase %d:", op+1)
		if strings.Count(text, start) != 1 || strings.Count(text, end) != 1 {
			t.Fatalf("nonunique score packet arm %d", op)
		}
		_, tail, _ := strings.Cut(text, start)
		body, _, _ := strings.Cut(tail, end)
		program = strings.Replace(program, fmt.Sprintf("// PRODUCTION_SCORE_ARM_%d", op), start+body, 1)
	}
	readers := ""
	if strings.Contains(text, "static uint16_t nox_cdecode_score_read_u16") {
		_, body, _ := strings.Cut(text, "static uint16_t nox_cdecode_score_read_u16")
		body, _, _ = strings.Cut(body, "//----- (0048EA70)")
		readers = "static uint16_t nox_cdecode_score_read_u16" + body
	}
	program = strings.Replace(program, "// PRODUCTION_SCORE_READERS", readers, 1)
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "client-score-packets")
	args := append(compiler[1:], "-std=c11", "-O2", "-fshort-wchar", "-fsigned-char",
		"-Wall", "-Wextra", "-Werror", "-Wno-sign-compare", "-Itestdata", "-x", "c", "-", "-o", bin)
	cmd := exec.Command(compiler[0], args...)
	cmd.Stdin = strings.NewReader(program)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile actual score packet arms: %v\n%s", err, out)
	}
	if out, err := exec.Command(bin).CombinedOutput(); err != nil {
		t.Fatalf("native score packet contracts: %v\n%s", err, out)
	} else {
		t.Log(strings.TrimSpace(string(out)))
	}
}
