//go:build !server

package opennox

import (
	"bytes"
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/cfg"
	"github.com/opennox/libs/env"
	"github.com/spf13/viper"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/client/seat/headless"
	"github.com/opennox/opennox/v1/legacy"
)

// Configuration/API fixture, not queued GUI input. A generated window layout
// is parsed by the actual constructor, which publishes its own native globals.
// No replacement constructor, option handler or video-mode hook is installed.
// Every case runs in a fresh subprocess and writes only its temporary cwd.
func TestOptionsInGameClose(t *testing.T) {
	for _, phase := range []string{"ingame-save", "ingame-cancel", "ingame-hidden", "ingame-e2e"} {
		t.Run(phase, func(t *testing.T) {
			runOptionsConfigChild(t, t.TempDir(), phase, 0)
		})
	}
}

func testOptionsInGameClose(t *testing.T, c *Client, sc *headless.Seat, phase string, _ int) {
	t.Helper()
	strMan = c.Strings()
	openedSize := image.Pt(800, 600)
	c.videoSetGameMode(openedSize)
	guiOptionsRes = image.Pt(640, 480) // A stale pending value before construction.
	if err := os.Mkdir("window", 0o700); err != nil {
		t.Fatal(err)
	}
	// Only independently generated geometry/control IDs, not stock window data,
	// images or fonts. The actual parser enhancement creates extended radios.
	var layout strings.Builder
	layout.WriteString("WINDOW\n300 0 0 640 423 USER;\nSTATUS = ENABLED+NOFOCUS;\nCHILD\n")
	for _, id := range []int{310, 311, 312, 313, 314, 320, 321, 322, 323, 330, 331, 332, 333, 334, 341, 361, 362, 363, 371} {
		fmt.Fprintf(&layout, "WINDOW\n%d 0 0 20 20 USER;\nSTATUS = ENABLED+NOFOCUS;\nEND\n", id)
	}
	for id := 351; id <= 353; id++ {
		fmt.Fprintf(&layout, "WINDOW\n%d 0 0 140 20 HORZSLIDER;\nSTATUS = ENABLED+NOFOCUS;\nDATA = 0 100;\nEND\n", id)
	}
	layout.WriteString("END\nEND\n")
	if err := os.WriteFile(filepath.Join("window", "Options.wnd"), []byte(layout.String()), 0o600); err != nil {
		t.Fatal(err)
	}
	// The real pause-menu owner is required by sub_445C40 after options close.
	// A visible C-owned root exercises its hide branch without unrelated game
	// startup. It is not a replacement for the queued pause-menu E2E.
	quit := c.GUI.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 20, 20, nil)
	quit.SetID(9000)
	legacy.Set_nox_wnd_quitMenu_825760(quit)
	t.Cleanup(func() { legacy.Set_nox_wnd_quitMenu_825760(nil) })
	if got := legacy.Nox_game_initOptionsInGame_4ADAD0(); got != 1 {
		t.Fatalf("actual in-game Options constructor returned %d", got)
	}
	root := c.GUI.ChildByID(300)
	if root == nil || !root.GetFlags().IsHidden() {
		t.Fatalf("constructor did not publish the hidden native Options root: %p", root)
	}
	if guiOptionsRes != openedSize || viper.GetInt(configVideoWidth) != openedSize.X || viper.GetInt(configVideoHeight) != openedSize.Y {
		t.Fatal("actual constructor did not synchronize the displayed and pending resolution")
	}
	for _, ptr := range []unsafe.Pointer{root.C(), root.ChildByID(351).C(), root.ChildByID(351).Field100().C(), quit.C()} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
			t.Fatalf("native fixture pointer %p did not exceed 4 GiB", ptr)
		}
	}
	initial := image.Pt(640, 480)
	const sentinel = "private in-game legacy sentinel\n"
	writeSentinel := func() {
		t.Helper()
		if err := os.WriteFile("nox.cfg", []byte(sentinel), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if phase == "ingame-save" {
		// Do not change the selected radio: Close must keep the opened size,
		// not apply the deliberately stale pre-constructor pending value.
		root.Show()
		quit.Show()
		c.UpdateFullScreen(-2)
		g_fullscreen_cfg = -3
		writeSentinel()
		close := root.ChildByID(371)
		if got := gui.EventRespInt(root.Func94(&gui.RawEvent{Event: 0x4007, Arg1: uintptr(close.C())})); got != 1 {
			t.Fatalf("unchanged-resolution C Close returned %d", got)
		}
		if sc.ScreenSize() != openedSize || c.videoGetGameMode() != openedSize || g_fullscreen_cfg != -2 || !root.GetFlags().IsHidden() {
			t.Fatalf("unchanged-resolution Close changed the size or missed its checkpoint: seat=%v game=%v checkpoint=%d", sc.ScreenSize(), c.videoGetGameMode(), g_fullscreen_cfg)
		}
		data, err := os.ReadFile("nox.cfg")
		if err != nil {
			t.Fatal(err)
		}
		file, err := cfg.Parse(bytes.NewReader(data))
		if err != nil || len(file.Sections) != 2 {
			t.Fatalf("unchanged-resolution Close wrote invalid config: %v", err)
		}
		if mode, ok := file.Sections[0].Get("VideoMode"); !ok || mode != "800 600 16" {
			t.Fatalf("unchanged-resolution Close wrote %q present=%t", mode, ok)
		}
	}
	for index, res := range getResolutionOptions() {
		if res == (image.Point{}) {
			continue
		}
		// API staging: reopening the real constructor-owned root is intentional;
		// mouse routing, draw-buffer rebuild and game-loop resize are separate.
		root.Show()
		quit.Show()
		c.videoSetGameMode(initial)
		c.UpdateFullScreen(-2) // Live borderless, different from the save checkpoint.
		g_fullscreen_cfg = -3
		writeSentinel()
		control := root.ChildByID(uint(guiIDMenuExt + index))
		if control == nil {
			t.Fatalf("constructor did not create resolution %d", index)
		}
		if got := gui.EventRespInt(root.Func94(&gui.RawEvent{Event: 0x4007, Arg1: uintptr(control.C())})); got != 1 {
			t.Fatalf("C resolution %d returned %d", index, got)
		}
		if guiOptionsRes != res || viper.GetInt(configVideoWidth) != res.X || viper.GetInt(configVideoHeight) != res.Y {
			t.Fatalf("C selection %d did not reach the real configuration handler", index)
		}
		if sc.ScreenSize() != initial || c.videoGetGameMode() != initial || g_fullscreen_cfg != -3 {
			t.Fatal("resolution selection applied before options close")
		}
		switch phase {
		case "ingame-cancel":
			legacy.Sub_4AD9B0(1) // Existing non-saving hide, e.g. observer transition.
		case "ingame-hidden":
			root.Hide()
			legacy.Sub_4AD9B0(0) // Already hidden is an original no-op.
		default:
			close := root.ChildByID(371)
			if got := gui.EventRespInt(root.Func94(&gui.RawEvent{Event: 0x4007, Arg1: uintptr(close.C())})); got != 1 {
				t.Fatalf("C Close returned %d", got)
			}
		}
		if !root.GetFlags().IsHidden() {
			t.Fatal("options close/hide left its native root visible")
		}
		wantMode, wantCheckpoint := initial, -3
		if phase == "ingame-save" {
			wantMode, wantCheckpoint = res, -2
		}
		if got := sc.ScreenSize(); got != wantMode || c.videoGetGameMode() != wantMode || g_fullscreen_cfg != wantCheckpoint {
			t.Errorf("%s resolution %d: seat=%v game=%v checkpoint=%d; want %v/%d", phase, index, got, c.videoGetGameMode(), g_fullscreen_cfg, wantMode, wantCheckpoint)
		}
		data, err := os.ReadFile("nox.cfg")
		if err != nil {
			t.Fatal(err)
		}
		if phase != "ingame-save" {
			if string(data) != sentinel {
				t.Fatal("non-saving/hidden/E2E close changed legacy disk config")
			}
			continue
		}
		file, err := cfg.Parse(bytes.NewReader(data))
		if err != nil || len(file.Sections) != 2 {
			t.Fatalf("actual in-game writer produced invalid config: %v", err)
		}
		if got, ok := file.Sections[0].Get("Fullscreen"); !ok || got != "-2" {
			t.Errorf("Close did not save the synchronized live window mode: %q present=%t", got, ok)
		}
		if got, ok := file.Sections[0].Get("VideoMode"); !ok || got != fmt.Sprintf("%d %d 16", res.X, res.Y) {
			t.Errorf("Close did not save the actual game dimensions: %q present=%t", got, ok)
		}
	}
	if env.IsE2E() != (phase == "ingame-e2e") {
		t.Fatal("wrong E2E isolation")
	}
}
