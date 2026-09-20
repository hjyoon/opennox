package gui

import (
	"math"
	"testing"
	"unsafe"
)

func TestProgressBarNativeParentAndValue(t *testing.T) {
	g := New(nil)
	defer g.alloc.Free()

	parent := g.NewWindowRaw(nil, StatusEnabled, 5, 7, 300, 200, nil)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(parent.C()) <= math.MaxUint32 {
		t.Fatalf("parent pointer = %p, want native address above 4 GiB", parent)
	}
	draw := WindowData{Window: parent, Style: StyleProgressBar}
	win := NewProgressBarRaw(g, parent, StatusEnabled|StatusSmoothText, 10, 20, 120, 16, &draw)
	if win == nil {
		t.Fatal("NewProgressBarRaw returned nil")
	}
	if win.Parent() != parent {
		t.Fatalf("parent = %p, want %p", win.Parent(), parent)
	}
	if win.GetFlags().Has(StatusEnabled) {
		t.Fatal("progress bar retained StatusEnabled")
	}
	if win.DrawData().Window != parent {
		t.Fatalf("draw owner = %p, want %p", win.DrawData().Window, parent)
	}
	if win.ext().Func94 == nil || win.ext().Draw == nil || win.field94 != nil || win.drawFunc != nil {
		t.Fatal("progress bar callbacks are not native Go callbacks")
	}

	win.Func94(AsWindowEvent(progressBarSetValueEvent, 42, 0))
	if got := progressBarValue(win); got != 42 {
		t.Fatalf("progress = %d, want 42", got)
	}
	win.Func94(AsWindowEvent(progressBarSetValueEvent, 101, 0))
	if got := progressBarValue(win); got != 42 {
		t.Fatalf("progress after 101 = %d, want 42", got)
	}
	win.Func94(AsWindowEvent(progressBarSetValueEvent, ^uintptr(0), 0))
	if got := progressBarValue(win); got != 42 {
		t.Fatalf("progress after -1 = %d, want 42", got)
	}
	win.Func94(AsWindowEvent(progressBarSetValueEvent, 100, 0))
	if got := progressBarValue(win); got != 100 {
		t.Fatalf("progress = %d, want 100", got)
	}
}

func TestProgressBarNativeRejectsWrongStyle(t *testing.T) {
	g := New(nil)
	defer g.alloc.Free()

	if win := NewProgressBarRaw(g, nil, 0, 0, 0, 100, 10, &WindowData{}); win != nil {
		t.Fatalf("NewProgressBarRaw returned %p for non-progress style", win)
	}
}
