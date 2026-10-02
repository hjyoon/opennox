package legacy

import (
	"image"
	"log/slog"
	"testing"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/client/input"
	seatheadless "github.com/opennox/opennox/v1/client/seat/headless"
)

type inputCfgCaptureFocusClient4CC170 struct {
	Client
	cli *client.Client
	sm  *strman.StringManager
}

func (c *inputCfgCaptureFocusClient4CC170) Cli() *client.Client            { return c.cli }
func (c *inputCfgCaptureFocusClient4CC170) Strings() *strman.StringManager { return c.sm }

// API fixture for the actual C capture procedure. It models the stock
// NOFOCUS root and detached modal without loading or changing any assets or
// configuration. Binding columns are deliberately absent; the full queued
// GUI audit checks real binding changes separately.
func newInputCfgCaptureFocusFixture4CC170(t *testing.T) (*gui.GUI, *gui.Window, *gui.Window, *gui.Window) {
	t.Helper()
	oldClient, oldBackground := GetClient, gui.MainBg
	oldPrompt := *inputCfgCapturePromptSlot4CC170
	oldSelection := *inputCfgCaptureSelectionSlot4CC170
	g := gui.New(nil)
	t.Cleanup(func() {
		g.DestroyAll()
		*inputCfgCapturePromptSlot4CC170 = oldPrompt
		*inputCfgCaptureSelectionSlot4CC170 = oldSelection
		GetClient, gui.MainBg = oldClient, oldBackground
	})
	c := &inputCfgCaptureFocusClient4CC170{cli: &client.Client{GUI: g}, sm: strman.New()}
	GetClient = func() Client { return c }
	acceptFocus := func(_ *gui.Window, ev gui.WindowEvent) gui.WindowEventResp {
		if _, ok := ev.(gui.WindowFocus); ok {
			return gui.RawEventResp(1)
		}
		return nil
	}
	background := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 640, 480, acceptFocus)
	root := g.NewWindowRaw(nil, gui.StatusEnabled|gui.StatusNoFocus, 0, 0, 640, 480, acceptFocus)
	root.SetID(900)
	prompt := g.NewWindowRaw(nil, gui.StatusEnabled, 10, 10, 80, 40, acceptFocus)
	prompt.SetID(980)
	for _, win := range []*gui.Window{background, root, prompt} {
		requireNativeWindowAddress(t, win)
	}
	gui.MainBg = background
	*inputCfgCapturePromptSlot4CC170 = prompt
	*inputCfgCaptureSelectionSlot4CC170 = nil
	prompt.SetFunc93C(inputCfgCapturePromptCEntry4CC170)
	if got := prompt.StackPush(); got != 0 {
		t.Fatalf("modal stack push: %d", got)
	}
	prompt.Focus()
	root.Focus()
	if g.Focused() != prompt || !root.Capture(true) {
		t.Fatal("capture fixture lost modal focus or did not retain NOFOCUS")
	}
	return g, root, prompt, background
}

func TestInputCfgCaptureCExitRestoresShellFocus(t *testing.T) {
	for _, tc := range []struct {
		name   string
		event  gui.WindowEvent
		closed bool
		result int
	}{
		{"left_down", gui.AsWindowEvent(6, 0, 0), true, 1},
		{"left_up", gui.AsWindowEvent(7, 0, 0), true, 1},
		{"right_down", gui.AsWindowEvent(10, 0, 0), true, 1},
		{"right_up", gui.AsWindowEvent(11, 0, 0), true, 1},
		{"middle_down", gui.AsWindowEvent(14, 0, 0), true, 1},
		{"middle_up", gui.AsWindowEvent(15, 0, 0), true, 1},
		{"wheel_up", gui.AsWindowEvent(19, 0, 0), true, 1},
		{"wheel_down", gui.AsWindowEvent(20, 0, 0), true, 1},
		{"escape_press", gui.WindowKeyPress{Key: keybind.KeyEsc, Pressed: true}, true, 1},
		{"escape_release", gui.WindowKeyPress{Key: keybind.KeyEsc, Pressed: false}, false, 1},
		{"valid_key_release", gui.WindowKeyPress{Key: keybind.KeyF9, Pressed: false}, true, 1},
		{"valid_key_press", gui.WindowKeyPress{Key: keybind.KeyF9, Pressed: true}, false, 0},
		{"invalid_key_release", gui.WindowKeyPress{Key: keybind.Key(0xffff), Pressed: false}, false, 0},
		{"other_event", gui.WindowNewChild{}, false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, root, prompt, background := newInputCfgCaptureFocusFixture4CC170(t)
			if got := gui.EventRespInt(prompt.Func93(tc.event)); got != tc.result {
				t.Fatalf("C callback response=%d, want %d", got, tc.result)
			}
			wantFocus, wantStack := prompt, prompt
			if tc.closed {
				wantFocus, wantStack = background, nil
			}
			if g.Focused() != wantFocus || g.StackHead() != wantStack || prompt.GetFlags().IsHidden() != tc.closed {
				t.Fatalf("closed=%t focus=%p want=%p stack=%p want=%p hidden=%t", tc.closed,
					g.Focused(), wantFocus, g.StackHead(), wantStack, prompt.GetFlags().IsHidden())
			}
			if g.Captured() != root || !root.GetFlags().Has(gui.StatusNoFocus) || *inputCfgCaptureSelectionSlot4CC170 != nil {
				t.Fatal("C exit changed unrelated capture/root flags/selection")
			}
		})
	}
}

func TestInputCfgCaptureQueuedKeysDoNotDismissParent(t *testing.T) {
	for _, cancel := range []bool{false, true} {
		t.Run(map[bool]string{false: "rebind", true: "cancel"}[cancel], func(t *testing.T) {
			g, root, prompt, background := newInputCfgCaptureFocusFixture4CC170(t)
			st := seatheadless.New(image.Pt(640, 480))
			t.Cleanup(func() { _ = st.Close() })
			h := input.New(slog.Default(), st, false, 0)
			h.SetDrawWinSize(image.Pt(640, 480))
			g.SetInput(h)
			var parentEscapes int
			background.SetFunc93(func(_ *gui.Window, ev gui.WindowEvent) gui.WindowEventResp {
				if key, ok := ev.(gui.WindowKeyPress); ok && key.Key == keybind.KeyEsc {
					if key.Pressed {
						parentEscapes++
					}
					return gui.RawEventResp(1)
				}
				return nil
			})
			send := func(key keybind.Key, pressed bool) {
				st.QueueInput(&seat.KeyboardEvent{Key: key, Pressed: pressed})
				h.Tick()
				g.ProcessKeys(h)
			}
			key := keybind.KeyF9
			if cancel {
				key = keybind.KeyEsc
			}
			send(key, true)
			if !cancel && (g.Focused() != prompt || prompt.GetFlags().IsHidden()) {
				t.Fatal("normal key press prematurely dismissed the capture prompt")
			}
			send(key, false)
			if g.Focused() != background || g.StackHead() != nil || !prompt.GetFlags().IsHidden() || parentEscapes != 0 {
				t.Fatalf("capture did not return keyboard input without dismissing parent: focus=%p want=%p stack=%p hidden=%t parent escapes=%d",
					g.Focused(), background, g.StackHead(), prompt.GetFlags().IsHidden(), parentEscapes)
			}
			send(keybind.KeyEsc, true)
			send(keybind.KeyEsc, false)
			if parentEscapes != 1 || g.Captured() != root || !root.GetFlags().Has(gui.StatusNoFocus) {
				t.Fatalf("next ESC did not reach the parent exactly once: escapes=%d capture=%p", parentEscapes, g.Captured())
			}
		})
	}
}
