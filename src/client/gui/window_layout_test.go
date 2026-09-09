package gui

import (
	"testing"
	"unsafe"
)

func TestWindowNativeLayout46C2A0(t *testing.T) {
	var win Window
	if got := unsafe.Offsetof(win.Flags); got != 4 {
		t.Fatalf("Window.Flags offset = %d, want 4", got)
	}

	switch ptrSize := unsafe.Sizeof(uintptr(0)); ptrSize {
	case 4:
		if got := unsafe.Sizeof(win); got != 404 {
			t.Fatalf("32-bit Window size = %d, want 404", got)
		}
		if got := unsafe.Offsetof(win.field93); got != 372 {
			t.Fatalf("32-bit Window.field93 offset = %d, want 372", got)
		}
		if got := unsafe.Offsetof(win.field94); got != 376 {
			t.Fatalf("32-bit Window.field94 offset = %d, want 376", got)
		}
		if got := unsafe.Offsetof(win.drawFunc); got != 380 {
			t.Fatalf("32-bit Window.drawFunc offset = %d, want 380", got)
		}
		if got := unsafe.Offsetof(win.parent); got != 396 {
			t.Fatalf("32-bit Window.parent offset = %d, want 396", got)
		}
	case 8:
		if got := unsafe.Sizeof(win); got != 528 {
			t.Fatalf("64-bit Window size = %d, want 528", got)
		}
		if got := unsafe.Offsetof(win.field93); got != 464 {
			t.Fatalf("64-bit Window.field93 offset = %d, want 464", got)
		}
		if got := unsafe.Offsetof(win.field94); got != 472 {
			t.Fatalf("64-bit Window.field94 offset = %d, want 472", got)
		}
		if got := unsafe.Offsetof(win.drawFunc); got != 480 {
			t.Fatalf("64-bit Window.drawFunc offset = %d, want 480", got)
		}
		if got := unsafe.Offsetof(win.parent); got != 512 {
			t.Fatalf("64-bit Window.parent offset = %d, want 512", got)
		}
	default:
		t.Fatalf("unsupported pointer size %d", ptrSize)
	}
}
