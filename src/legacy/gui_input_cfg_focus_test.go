package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

// API fixture for the new C-to-Go primitive, not queued input or a claim that
// the shell capture callback uses it yet. No asset/config file is required.
func TestInputCfgRestoreFocusCEntry(t *testing.T) {
	oldBackground := gui.MainBg
	t.Cleanup(func() { gui.MainBg = oldBackground })
	g := gui.New(nil)
	t.Cleanup(g.DestroyAll)
	var backgroundFocus, promptBlur int
	background := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 640, 480, func(_ *gui.Window, ev gui.WindowEvent) gui.WindowEventResp {
		if focus, ok := ev.(gui.WindowFocus); ok {
			if focus {
				backgroundFocus++
			}
			return gui.RawEventResp(1)
		}
		return nil
	})
	prompt := g.NewWindowRaw(nil, gui.StatusEnabled, 10, 10, 80, 40, func(_ *gui.Window, ev gui.WindowEvent) gui.WindowEventResp {
		if focus, ok := ev.(gui.WindowFocus); ok {
			if !focus {
				promptBlur++
			}
			return gui.RawEventResp(1)
		}
		return nil
	})
	requireNativeWindowAddress(t, background)
	requireNativeWindowAddress(t, prompt)
	gui.MainBg = background
	if got := prompt.StackPush(); got != 0 {
		t.Fatalf("modal stack push: %d", got)
	}
	prompt.Focus()
	ccall.CallVoidVoid(inputCfgRestoreFocusCEntry)
	if g.Focused() != background || backgroundFocus != 1 || promptBlur != 1 {
		t.Fatalf("C focus restore missed the native background: focused=%p background=%p accepted=%d blurred=%d", g.Focused(), background, backgroundFocus, promptBlur)
	}
	if g.StackHead() != prompt || g.Captured() != nil || prompt.GetFlags().IsHidden() {
		t.Fatal("focus primitive changed modal stack/capture/visibility")
	}
	// The generic nil-focus API remains a clear, not a shell fallback.
	g.Focus(nil)
	if g.Focused() != nil {
		t.Fatal("generic nil focus no longer clears focus")
	}
	for _, unavailable := range []string{"nil", "destroyed"} {
		gui.MainBg = nil
		if unavailable == "destroyed" {
			dead := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 10, 10, nil)
			dead.Destroy()
			gui.MainBg = dead
		}
		ccall.CallVoidVoid(inputCfgRestoreFocusCEntry)
		if g.Focused() != nil {
			t.Fatalf("%s background changed focus", unavailable)
		}
	}
	if got := prompt.StackPop(); got != 0 {
		t.Fatalf("modal stack pop: %d", got)
	}
}
