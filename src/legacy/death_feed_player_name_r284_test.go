package legacy

import (
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"golang.org/x/image/font"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type deathFeedEntryTestRender struct {
	Render2
	entry *deathFeedEntry
}

func (r deathFeedEntryTestRender) DrawString(face font.Face, text string, point image.Point) int {
	result := r.Render2.DrawString(face, text, point)
	r.entry.victim = 0
	return result
}

type deathFeedEntryTestClient struct {
	Client
	render Render2
}

func (c deathFeedEntryTestClient) R2() Render2 { return c.render }

func TestDeathFeed495500NativeWordThroughC(t *testing.T) {
	g := newWidgetCallbackGUI(t)
	defer g.DestroyAll()
	data := GetClient().R2().Data()
	data.SetClip(true)
	data.SetClipRect(image.Rect(0, 0, 640, 480))

	oldWidth, oldRow, oldIcon := clientDisplayStorage.width, legacyWords.dword_5d4594_1203832, deathFeedFallbackIcon
	clientDisplayStorage.width, legacyWords.dword_5d4594_1203832, deathFeedFallbackIcon = -50, 0, nil
	defer func() {
		clientDisplayStorage.width, legacyWords.dword_5d4594_1203832, deathFeedFallbackIcon = oldWidth, oldRow, oldIcon
	}()
	entry, free := alloc.NewHigh(deathFeedEntry{})
	defer free()
	address := uintptr(unsafe.Pointer(entry))
	if unsafe.Sizeof(address) > 4 && address <= uintptr(^uint32(0)) {
		t.Fatalf("entry address %#x does not exercise the native width", address)
	}
	if got := deathFeedEntryCallC(entry); got != address {
		t.Fatalf("no victim returned %#x, want entry address %#x", got, address)
	}
	entry.victim = 1
	clientDisplayStorage.width = 640
	if got := deathFeedEntryCallC(entry); got != 325 {
		t.Fatalf("victim text returned %#x, want positive coordinate 325", got)
	}

	clientDisplayStorage.width = -50
	oldClient := GetClient
	base := GetClient()
	GetClient = func() Client {
		return deathFeedEntryTestClient{Client: base, render: deathFeedEntryTestRender{Render2: base.R2(), entry: entry}}
	}
	defer func() { GetClient = oldClient }()
	if got, want := deathFeedEntryCallC(entry), ^uintptr(0)-19; got != want || entry.victim != 0 {
		t.Fatalf("victim text returned %#x with victim %d, want signed coordinate %#x and cleared victim", got, entry.victim, want)
	}
}

// The retired death-feed consumer is compiled with actual native
// headers and UTF-16 formatting. Lookup/font/Thing/image/draw services are
// controlled; no linked renderer, original executable or GUI is exercised.
func TestDeathFeedPlayerNamesNativeR284(t *testing.T) {
	source, err := os.ReadFile("testdata/retired-game-sources/death-feed-495210.c.txt")
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
