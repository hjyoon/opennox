package legacy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Compile the three complete production C consumers against the real native
// Player, chat-bubble and window headers. Font/layout/draw/lookup/team/window
// services are controlled; this is not a linked GUI, font or network test.
func clientPlayerNameNativeR284(t *testing.T, mode string) {
	t.Helper()
	fixture, err := os.ReadFile("testdata/client_player_name_r284_test.c")
	if err != nil {
		t.Fatal(err)
	}
	program := string(fixture)
	for _, part := range []struct{ file, start, end, marker string }{
		{"GAME2_3.c", "void sub_48D990(nox_draw_viewport_t* a1p) {", "//----- (0048DCF0)", "// PRODUCTION_CHAT_NAME"},
		{"GAME3_1.c", "int nox_xxx_netP2PStartTrade_4C1320(const uint8_t* data) {", "//----- (004C1410)", "// PRODUCTION_TRADE_START_NAME"},
		{"GAME3_1.c", "int nox_xxx_prepareP2PTrade_4C1BF0() {", "//----- (004C1CA0)", "// PRODUCTION_TRADE_PREPARE_NAME"},
	} {
		source, err := os.ReadFile(part.file)
		if err != nil {
			t.Fatal(err)
		}
		text := string(source)
		if strings.Count(text, part.start) != 1 || strings.Count(text, part.end) != 1 || strings.Count(program, part.marker) != 1 {
			t.Fatalf("ambiguous complete production boundary: %s", part.start)
		}
		_, tail, _ := strings.Cut(text, part.start)
		body, _, ok := strings.Cut(tail, part.end)
		if !ok || !strings.HasSuffix(strings.TrimSpace(body), "}") {
			t.Fatalf("incomplete production body: %s", part.start)
		}
		program = strings.Replace(program, part.marker, part.start+body, 1)
	}
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "client-player-name-r284")
	args := append(compiler[1:], "-std=c11", "-O2", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing",
		"-Wall", "-Wextra", "-Werror", "-Itestdata", "-x", "c", "-", "-o", bin)
	cmd := exec.Command(compiler[0], nativeFixtureCompileArgs(t, args)...)
	cmd.Stdin = strings.NewReader(program)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile complete native name consumers: %v\n%s", err, out)
	}
	if out, err := exec.Command(bin, mode).CombinedOutput(); err != nil {
		t.Fatalf("native %s name consumer: %v\n%s", mode, err, out)
	} else {
		t.Log(strings.TrimSpace(string(out)))
	}
}

func TestChatPlayerNameNativeR284(t *testing.T) {
	clientPlayerNameNativeR284(t, "chat")
}

func TestP2PStartTradePlayerNameNativeR284(t *testing.T) {
	clientPlayerNameNativeR284(t, "start")
}

func TestP2PPrepareTradePlayerNameNativeR284(t *testing.T) {
	clientPlayerNameNativeR284(t, "prepare")
}
