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

// Compile complete actual decoder arms and their Player declarations. Service
// callbacks are controlled, while formatting uses the real UTF-16 code.
func clientQuestNoticesNative48EA70(t *testing.T, mode string) {
	t.Helper()
	source, err := os.ReadFile("client__network__cdecode.c")
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("testdata/client_quest_notices_48ea70_test.c")
	if err != nil {
		t.Fatal(err)
	}
	text, program := string(source), string(fixture)
	for _, name := range []string{"v303", "v306", "v309"} {
		matches := regexp.MustCompile(`(?m)^\t(?:char\*|nox_playerInfo\*) `+name+`;[^\n]*`).FindAllString(text, -1)
		if len(matches) != 1 {
			t.Fatalf("ambiguous Player declaration: %s", name)
		}
		program = strings.Replace(program, "// PLAYER_"+name, matches[0], 1)
	}
	_, quest, ok := strings.Cut(text, "\tcase 0xF0u:")
	if !ok {
		t.Fatal("missing actual Quest arm")
	}
	for _, sub := range []int{30, 31, 32} {
		start, end := fmt.Sprintf("\t\tcase 0x%Xu:", sub), fmt.Sprintf("\t\tcase 0x%Xu:", sub+1)
		if strings.Count(quest, start) != 1 || strings.Count(quest, end) != 1 {
			t.Fatalf("ambiguous Quest subcommand %d", sub)
		}
		_, tail, _ := strings.Cut(quest, start)
		body, _, _ := strings.Cut(tail, end)
		program = strings.Replace(program, fmt.Sprintf("// QUEST_ARM_%d", sub), start+body, 1)
	}
	start, end := "\tcase 0xB1u:", "\tcase 0xB2u:"
	if strings.Count(text, start) != 1 || strings.Count(text, end) != 1 {
		t.Fatal("ambiguous settings/name arm")
	}
	_, tail, _ := strings.Cut(text, start)
	body, _, _ := strings.Cut(tail, end)
	program = strings.Replace(program, "// NAME_CHANGE_ARM", start+body, 1)
	_, readers, ok := strings.Cut(text, "static uint16_t nox_cdecode_score_read_u16")
	if !ok {
		t.Fatal("missing actual packed wire readers")
	}
	readers, _, _ = strings.Cut(readers, "//----- (0048EA70)")
	program = strings.Replace(program, "// WIRE_READERS", "static uint16_t nox_cdecode_score_read_u16"+readers, 1)
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "client-quest-notices")
	args := append(compiler[1:], "-std=c11", "-O2", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing",
		"-Wall", "-Wextra", "-Werror", "-Wno-unused-function", "-Itestdata", "-x", "c", "-", "-o", bin)
	cmd := exec.Command(compiler[0], nativeFixtureCompileArgs(t, args)...)
	cmd.Stdin = strings.NewReader(program)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile actual Quest/settings arms: %v\n%s", err, out)
	}
	if out, err := exec.Command(bin, mode).CombinedOutput(); err != nil {
		t.Fatalf("actual %s contracts: %v\n%s", mode, err, out)
	} else {
		t.Log(strings.TrimSpace(string(out)))
	}
}

func TestClientQuestRewards48EA70NativeBodies(t *testing.T) {
	clientQuestNoticesNative48EA70(t, "rewards")
}

func TestClientNameChange48EA70NativeBody(t *testing.T) {
	clientQuestNoticesNative48EA70(t, "settings")
}
