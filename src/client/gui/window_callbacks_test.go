package gui

import (
	"testing"
	"unsafe"
)

func TestWindowNativeCallbackStorage(t *testing.T) {
	g := New(nil)
	defer g.alloc.Free()

	win := g.NewWindowRaw(nil, StatusEnabled, 0, 0, 100, 100, nil)
	var proc93, proc94, draw byte

	win.SetFunc93C(unsafe.Pointer(&proc93))
	if win.field93 != unsafe.Pointer(&proc93) || win.ext().Func93 != nil {
		t.Fatal("SetFunc93C did not retain the native callback")
	}
	win.SetFunc93(nil)
	if win.field93 != nil || win.ext().Func93 == nil {
		t.Fatal("SetFunc93 did not replace the native callback")
	}

	win.SetFunc94C(unsafe.Pointer(&proc94))
	if win.field94 != unsafe.Pointer(&proc94) || win.ext().Func94 != nil {
		t.Fatal("SetFunc94C did not retain the native callback")
	}
	win.SetFunc94(nil)
	if win.field94 != nil || win.ext().Func94 == nil {
		t.Fatal("SetFunc94 did not replace the native callback")
	}

	win.SetDrawC(unsafe.Pointer(&draw))
	if win.drawFunc != unsafe.Pointer(&draw) || win.ext().Draw != nil {
		t.Fatal("SetDrawC did not retain the native callback")
	}
	win.SetDraw(nil)
	if win.drawFunc != nil || win.ext().Draw == nil {
		t.Fatal("SetDraw did not replace the native callback")
	}
}

func TestResolveLegacyWindow(t *testing.T) {
	g := New(nil)
	defer g.alloc.Free()

	first := g.NewWindowRaw(nil, StatusEnabled, 0, 0, 100, 100, nil)
	second := g.NewWindowRaw(nil, StatusEnabled, 0, 0, 100, 100, nil)
	defer setExt(first, nil)
	defer setExt(second, nil)

	firstAddr := uintptr(unsafe.Pointer(first))
	if got := g.ResolveLegacyWindow(firstAddr); got != first {
		t.Fatalf("exact pointer resolved to %p, want %p", got, first)
	}
	if got := g.ResolveLegacyWindow(firstAddr ^ 1); got != nil {
		t.Fatalf("unknown pointer resolved to %p", got)
	}

	if unsafe.Sizeof(firstAddr) <= 4 {
		return
	}
	if firstAddr <= uintptr(^uint32(0)) {
		t.Fatalf("test window %p does not exercise native pointer truncation", first)
	}
	low := uint32(firstAddr)
	if got := g.ResolveLegacyWindow(uintptr(low)); got != first {
		t.Fatalf("zero-extended alias %#x resolved to %p, want %p", low, got, first)
	}
	signed := uintptr(uint64(int64(int32(low))))
	if got := g.ResolveLegacyWindow(signed); got != first {
		t.Fatalf("sign-extended alias %#x resolved to %p, want %p", signed, got, first)
	}
	wrongHighWord := uint32(0x12345678)
	if wrongHighWord == uint32(firstAddr>>32) {
		wrongHighWord = 0x23456789
	}
	wrongHigh := (uintptr(wrongHighWord) << 32) | uintptr(low)
	if got := g.ResolveLegacyWindow(wrongHigh); got != nil {
		t.Fatalf("non-legacy alias %#x resolved to %p", wrongHigh, got)
	}

	setExt(first, nil)
	if got := g.ResolveLegacyWindow(uintptr(low)); got != nil {
		t.Fatalf("deallocated alias %#x resolved to %p", low, got)
	}
}
