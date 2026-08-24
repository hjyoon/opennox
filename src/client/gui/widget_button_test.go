package gui

import (
	"image"
	"testing"

	"github.com/opennox/opennox/v1/client/input"
)

func TestButtonHoverAndClickAnimationState(t *testing.T) {
	g := New(nil)
	defer g.alloc.Free()

	draw := WindowData{Style: StyleMouseTrack}
	btn := NewButtonRaw(g, nil, StatusEnabled|StatusImage, 157, 91, 166, 35, &draw)
	if btn == nil {
		t.Fatal("NewButtonRaw returned nil")
	}

	ButtonProc2(btn, &WindowMouseUnk{Event: 17, Pos: image.Pt(240, 108)})
	if btn.DrawData().Field0&0x2 == 0 {
		t.Fatal("mouse enter did not select the highlight image state")
	}
	if g.Focused() != btn {
		t.Fatal("mouse enter did not focus the button")
	}

	ButtonProc2(btn, &WindowMouseState{State: input.NOX_MOUSE_LEFT_DOWN, Pos: image.Pt(240, 108)})
	if btn.DrawData().Field0&0x4 == 0 {
		t.Fatal("mouse down did not select the pressed image state")
	}

	ButtonProc2(btn, &WindowMouseState{State: input.NOX_MOUSE_LEFT_UP, Pos: image.Pt(240, 108)})
	if btn.DrawData().Field0&0x4 != 0 {
		t.Fatal("mouse up did not clear the pressed image state")
	}

	ButtonProc2(btn, &WindowMouseUnk{Event: 18, Pos: image.Pt(400, 108)})
	if btn.DrawData().Field0&0x2 != 0 {
		t.Fatal("mouse leave did not clear the highlight image state")
	}
}
