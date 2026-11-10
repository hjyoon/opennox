package gui

import (
	"testing"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/opennox/v1/client/input"
)

func TestRadioButtonReentrantKeyPreservesOriginalToggle(t *testing.T) {
	g := New(nil)
	defer func() { g.DestroyAll(); g.alloc.Free() }()
	var clicked *Window
	parent := g.NewWindowRaw(nil, StatusEnabled, 0, 0, 100, 100, func(_ *Window, e WindowEvent) WindowEventResp {
		if e.EventCode() == 0x4007 {
			clicked.DrawData().Field0 |= 4
		}
		return RawEventResp(1)
	})
	clicked = g.NewWindowRaw(parent, StatusEnabled, 0, 0, 20, 20, RadioButtonProcPre)
	d := WindowData{Window: parent, Style: StyleRadioButton}
	clicked.CopyDrawData(&d)
	// Original 4A84E0 tests the old bit, notifies, then XORs the live bit. An
	// owner that selects during notification must therefore end deselected.
	if EventRespInt(RadioButtonProcNative(clicked, WindowKeyPress{Key: keybind.KeyEnter, Pressed: true})) != 1 || clicked.DrawData().Field0&4 != 0 {
		t.Fatal("reentrant key notification lost the original live XOR")
	}
}

func TestRadioButtonReentrantNotifyKeepsOriginalParent(t *testing.T) {
	for _, pre := range []bool{false, true} {
		t.Run(map[bool]string{false: "mouse", true: "programmatic"}[pre], func(t *testing.T) {
			g := New(nil)
			defer func() { g.DestroyAll(); g.alloc.Free() }()
			var clicked, newParent *Window
			oldParent := g.NewWindowRaw(nil, StatusEnabled, 0, 0, 100, 100, func(_ *Window, e WindowEvent) WindowEventResp {
				if e.EventCode() == 0x4007 {
					g.setParent(clicked, newParent)
				}
				return RawEventResp(1)
			})
			newParent = g.NewWindowRaw(nil, StatusEnabled, 0, 0, 100, 100, nil)
			makeRadio := func(parent *Window) *Window {
				w := g.NewWindowRaw(parent, StatusEnabled, 0, 0, 20, 20, RadioButtonProcPre)
				d := WindowData{Window: oldParent, Style: StyleRadioButton}
				d.SetGroup(9)
				w.CopyDrawData(&d)
				return w
			}
			oldSibling := makeRadio(oldParent)
			newSibling := makeRadio(newParent)
			clicked = makeRadio(oldParent)
			oldSibling.DrawData().Field0 |= 4
			newSibling.DrawData().Field0 |= 4
			if pre {
				RadioButtonProcPre(clicked, AsWindowEvent(0x4008, 1, 0))
			} else {
				RadioButtonProcNative(clicked, &WindowMouseState{State: input.NOX_MOUSE_LEFT_UP})
			}
			if clicked.Parent() != newParent || oldSibling.DrawData().Field0&4 != 0 || newSibling.DrawData().Field0&4 == 0 || clicked.DrawData().Field0&4 == 0 {
				t.Fatal("notification reparent changed the original cached sibling traversal")
			}
		})
	}
}
