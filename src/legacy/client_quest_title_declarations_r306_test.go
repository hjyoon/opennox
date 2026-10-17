package legacy

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
	"unsafe"
)

// Use the same complete arms, Player declarations, wire readers and fixture as
// clientQuestNoticesNative48EA70, additionally binding all six title locals.
// The fixture's handwritten pointer locals otherwise hide production truncation.
// Its runtime varargs checks, UTF-16 output, helper order and whole-buffer checks
// remain unchanged, including all 2160 reward and 512 settings inputs.
func clientQuestNoticesProgramR306(t *testing.T) string {
	t.Helper()
	source, err := os.ReadFile("client__network__cdecode.c")
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("testdata/client_quest_notices_48ea70_test.c")
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ReplaceAll(string(source), "\r\n", "\n")
	program := strings.ReplaceAll(string(fixture), "\r\n", "\n")
	replace := func(marker, body string) {
		t.Helper()
		if strings.Count(program, marker) != 1 {
			t.Fatalf("nonunique Quest fixture insertion marker %q", marker)
		}
		program = strings.Replace(program, marker, body, 1)
	}
	declaration := func(name, types string) string {
		t.Helper()
		matches := regexp.MustCompile(`(?m)^\t(?:`+types+`) `+name+`;[^\n]*`).FindAllString(text, -1)
		if len(matches) != 1 {
			t.Fatalf("nonunique production Quest declaration %s", name)
		}
		return matches[0]
	}
	for _, name := range []string{"v303", "v306", "v309"} {
		replace("// PLAYER_"+name, declaration(name, `char\*|nox_playerInfo\*`))
	}
	var titles []string
	for i := 348; i <= 353; i++ {
		titles = append(titles, declaration(fmt.Sprintf("v%d", i), `int|wchar2_t\*`))
	}
	replace("\twchar2_t *v348,*v349,*v350,*v351,*v352,*v353;", strings.Join(titles, "\n"))
	if strings.Count(text, "\tcase 0xF0u:") != 1 {
		t.Fatal("nonunique production Quest arm")
	}
	_, quest, _ := strings.Cut(text, "\tcase 0xF0u:")
	for _, sub := range []int{30, 31, 32} {
		start, end := fmt.Sprintf("\t\tcase 0x%Xu:", sub), fmt.Sprintf("\t\tcase 0x%Xu:", sub+1)
		if strings.Count(quest, start) != 1 || strings.Count(quest, end) != 1 {
			t.Fatalf("nonunique production Quest subcommand %d", sub)
		}
		_, tail, _ := strings.Cut(quest, start)
		body, _, _ := strings.Cut(tail, end)
		replace(fmt.Sprintf("// QUEST_ARM_%d", sub), start+body)
	}
	start, end := "\tcase 0xB1u:", "\tcase 0xB2u:"
	if strings.Count(text, start) != 1 || strings.Count(text, end) != 1 {
		t.Fatal("nonunique production settings/name arm")
	}
	_, tail, _ := strings.Cut(text, start)
	body, _, _ := strings.Cut(tail, end)
	replace("// NAME_CHANGE_ARM", start+body)
	start, end = "static uint16_t nox_cdecode_score_read_u16", "//----- (0048EA70)"
	if strings.Count(text, start) != 1 || strings.Count(text, end) != 1 {
		t.Fatal("nonunique production packed-wire readers")
	}
	_, readers, _ := strings.Cut(text, start)
	readers, _, ok := strings.Cut(readers, end)
	if !ok {
		t.Fatal("missing production packed-wire reader boundary")
	}
	replace("// WIRE_READERS", start+readers)
	return program
}

func clientQuestNoticesEnvironmentR306() []string {
	env := make([]string, 0, len(os.Environ())+2)
	for _, value := range os.Environ() {
		if !strings.HasPrefix(value, "ASAN_OPTIONS=") && !strings.HasPrefix(value, "UBSAN_OPTIONS=") {
			env = append(env, value)
		}
	}
	// LeakSanitizer is unsupported on Darwin; address and UB checks stay fatal.
	return append(env, "ASAN_OPTIONS=detect_leaks=0:halt_on_error=1:abort_on_error=0",
		"UBSAN_OPTIONS=halt_on_error=1:print_stacktrace=1")
}

func TestClientQuestNotices48EA70ProductionTitleDeclarationsR306(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("64-bit production Quest title pointer regression")
	}
	program := clientQuestNoticesProgramR306(t)
	modes := []struct {
		name  string
		flags []string
	}{
		{"O0", []string{"-O0"}},
		{"O2", []string{"-O2"}},
		{"sanitized", []string{"-O1", "-fsanitize=address,undefined", "-fno-sanitize-recover=all", "-fno-omit-frame-pointer"}},
	}
	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			if mode.name == "sanitized" && runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
				t.Skip("sanitizer mode requires the supported Darwin/Linux C runtime")
			}
			compiler := strings.Fields(os.Getenv("CC"))
			if len(compiler) == 0 {
				compiler = []string{"cc"}
			}
			bin := filepath.Join(t.TempDir(), "client-quest-title-declarations")
			if runtime.GOOS == "windows" {
				bin += ".exe"
			}
			args := append([]string(nil), compiler[1:]...)
			args = append(args, "-std=c11", "-g", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing",
				"-Wall", "-Wextra", "-Werror", "-Wno-unused-function")
			// Keep the baseline int locals compilable so the original fixture's
			// varargs/payload assertions expose truncation at runtime, also at O0.
			args = append(args, "-Wno-int-conversion")
			args = append(args, mode.flags...)
			// Current Darwin SDK marks the unchanged fixture's decimal/radix
			// formatter deprecated when ASan enables its fortified declaration.
			// Retain that formatter and every runtime check in this regression.
			if runtime.GOOS == "darwin" {
				args = append(args, "-Wno-deprecated-declarations")
			}
			args = append(args, "-Itestdata", "-x", "c", "-", "-o", bin)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			cmd := exec.CommandContext(ctx, compiler[0], args...)
			cmd.Stdin = strings.NewReader(program)
			cmd.Env = clientQuestNoticesEnvironmentR306()
			out, err := cmd.CombinedOutput()
			cancel()
			if err != nil {
				t.Fatalf("compile production Quest title declarations: %v\n%s", err, out)
			}
			for _, contract := range []struct {
				name  string
				cases int
			}{{"rewards", 2160}, {"settings", 512}} {
				t.Run(contract.name, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					cmd := exec.CommandContext(ctx, bin, contract.name)
					cmd.Env = clientQuestNoticesEnvironmentR306()
					out, err := cmd.CombinedOutput()
					if err != nil {
						t.Fatalf("production Quest title varargs/payload contract: %v\n%s", err, out)
					}
					want := fmt.Sprintf("Quest notices %s: %d native cases;", contract.name, contract.cases)
					if !strings.Contains(string(out), want) {
						t.Fatalf("missing complete original fixture result %q: %s", want, out)
					}
					t.Log(strings.TrimSpace(string(out)))
				})
			}
		})
	}
}
