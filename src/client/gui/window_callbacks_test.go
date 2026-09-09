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
