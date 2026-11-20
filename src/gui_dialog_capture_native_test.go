package opennox

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/legacy"
)

func TestGame1SplitDialogRestoresCaptureAndFocus(t *testing.T) {
	for _, tc := range []struct {
		name       string
		dead       bool
		freed      bool
		alreadyCap bool
	}{
		{name: "live-prior-window"},
		{name: "queued-for-free", dead: true},
		{name: "freed-prior-window", dead: true, freed: true},
		{name: "newer-capture", alreadyCap: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			oldClient := noxClient
			oldDialog, oldOwner := nox_gui_curDialog_830224, dword_5d4594_830228
			oldFocus, oldCapture := dword_5d4594_830232, dword_5d4594_830236
			t.Cleanup(func() {
				noxClient = oldClient
				nox_gui_curDialog_830224, dword_5d4594_830228 = oldDialog, oldOwner
				dword_5d4594_830232, dword_5d4594_830236 = oldFocus, oldCapture
			})

			g := gui.NewHigh(nil)
			t.Cleanup(g.DestroyAll)
			noxClient = &Client{Client: &client.Client{GUI: g, Cursor: gui.CursorPickup}}
			prior := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 64, 48, nil)
			dialog := g.NewWindowRaw(nil, gui.StatusEnabled, 8, 8, 32, 24, nil)
			if prior == nil || dialog == nil {
				t.Fatal("could not allocate native windows")
			}
			if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(prior)) <= 0xffffffff || uintptr(unsafe.Pointer(dialog)) <= 0xffffffff) {
				t.Fatalf("window addresses must exceed 4 GiB: prior=%p dialog=%p", prior, dialog)
			}
			focusAfterDestroy := false
			prior.SetFunc94(func(_ *gui.Window, ev gui.WindowEvent) gui.WindowEventResp {
				if change, ok := ev.(gui.WindowFocus); ok && bool(change) {
					focusAfterDestroy = dialog.Flags.Has(gui.StatusDestroyed)
					return gui.RawEventResp(1)
				}
				return nil
			})

			// Reproduce the capture release and stack state of a newly opened dialog.
			if !prior.Capture(true) || dialog.StackPush() != 0 {
				t.Fatal("could not prepare a captured window and live dialog")
			}
			nox_gui_curDialog_830224, dword_5d4594_830228 = dialog, prior
			dword_5d4594_830232, dword_5d4594_830236 = prior, g.Captured()
			prior.Capture(false)
			wantCapture, wantFocus := prior, prior
			if tc.dead {
				other := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 16, 16, nil)
				other.SetFunc94(func(_ *gui.Window, ev gui.WindowEvent) gui.WindowEventResp {
					if _, ok := ev.(gui.WindowFocus); ok {
						return gui.RawEventResp(1)
					}
					return nil
				})
				g.Focus(other)
				prior.Destroy()
				if tc.freed {
					g.FreeDestroyed()
				}
				if g.OwnsWindow(prior) == tc.freed {
					t.Fatal("fixture did not produce the expected destroyed-window lifetime")
				}
				wantCapture, wantFocus = nil, other
			}
			if tc.alreadyCap {
				next := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 16, 16, nil)
				if !next.Capture(true) {
					t.Fatal("could not install a newer capture")
				}
				wantCapture = next
			}

			legacy.Sub_44A400()

			if got := g.Captured(); got != wantCapture {
				t.Errorf("captured window = %p, want %p", got, wantCapture)
			}
			if got := g.Focused(); got != wantFocus || (!tc.dead && !focusAfterDestroy) {
				t.Errorf("focused window = %p, want %p after destroying dialog, order=%t", got, wantFocus, focusAfterDestroy)
			}
			if !dialog.Flags.Has(gui.StatusDestroyed) || nox_gui_curDialog_830224 != nil ||
				dword_5d4594_830228 != nil || dword_5d4594_830236 != nil || noxClient.Cursor != gui.CursorSelect {
				t.Errorf("close did not clear dialog state or restore cursor: dialog=%p owner=%p saved capture=%p cursor=%d",
					nox_gui_curDialog_830224, dword_5d4594_830228, dword_5d4594_830236, noxClient.Cursor)
			}
		})
	}
}
