package gui

import (
	"image"
	"testing"

	"github.com/opennox/opennox/v1/client/input"
)

func TestRadioButtonNativeSelection(t *testing.T) {
	g := New(nil)
	defer g.alloc.Free()

	var events []int
	parent := g.NewWindowRaw(nil, StatusEnabled, 0, 0, 300, 200, func(_ *Window, e WindowEvent) WindowEventResp {
		events = append(events, e.EventCode())
		return RawEventResp(1)
	})
	newRadio := func(group int) *Window {
		draw := WindowData{Window: parent, Style: StyleRadioButton | StyleMouseTrack}
		draw.SetGroup(group)
		win := g.NewWindowRaw(parent, StatusEnabled|StatusImage, 0, 0, 100, 20, RadioButtonProcPre)
		win.CopyDrawData(&draw)
		RadioButtonInit(win)
		return win
	}
	a := newRadio(7)
	b := newRadio(7)
	c := newRadio(8)
	c.DrawData().Field0 |= 0x4

	radioButtonProc(a, &WindowMouseUnk{Event: 17, Pos: image.Pt(5, 5)})
	if a.DrawData().Field0&0x2 == 0 {
		t.Fatal("mouse enter did not set the highlight state")
	}
	radioButtonProc(a, &WindowMouseState{State: input.NOX_MOUSE_LEFT_UP, Pos: image.Pt(5, 5)})
	if a.DrawData().Field0&0x4 == 0 {
		t.Fatal("first radio button was not selected")
	}
	if c.DrawData().Field0&0x4 == 0 {
		t.Fatal("selecting a radio button changed another group")
	}

	radioButtonProc(b, &WindowMouseState{State: input.NOX_MOUSE_LEFT_UP, Pos: image.Pt(5, 5)})
	if b.DrawData().Field0&0x4 == 0 {
		t.Fatal("second radio button was not selected")
	}
	if a.DrawData().Field0&0x4 != 0 {
		t.Fatal("first radio button in the same group remained selected")
	}
	if len(events) == 0 || events[len(events)-1] != 0x4007 {
		t.Fatalf("parent events = %v, want final click event 0x4007", events)
	}
}

func TestRadioButtonProgrammaticSelection(t *testing.T) {
	g := New(nil)
	defer g.alloc.Free()

	clicks := 0
	parent := g.NewWindowRaw(nil, StatusEnabled, 0, 0, 100, 100, func(_ *Window, e WindowEvent) WindowEventResp {
		if e.EventCode() == 0x4007 {
			clicks++
		}
		return RawEventResp(1)
	})
	draw := WindowData{Window: parent, Style: StyleRadioButton}
	win := g.NewWindowRaw(parent, StatusEnabled, 0, 0, 20, 20, RadioButtonProcPre)
	win.CopyDrawData(&draw)

	RadioButtonProcPre(win, AsWindowEvent(0x4008, 1, 0))
	if win.DrawData().Field0&0x4 == 0 {
		t.Fatal("0x4008 did not select the radio button")
	}
	if clicks != 1 {
		t.Fatalf("click notifications = %d, want 1", clicks)
	}
}
