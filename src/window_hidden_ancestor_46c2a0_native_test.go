package opennox

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
)

func TestWindowHiddenAncestorNative46C2A0HighPointersAndLowByteFlags(t *testing.T) {
	g := gui.New(nil)
	t.Cleanup(g.DestroyAll)

	root := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 100, 100, nil)
	parent := g.NewWindowRaw(root, gui.StatusEnabled, 0, 0, 80, 80, nil)
	win := g.NewWindowRaw(parent, gui.StatusEnabled, 0, 0, 40, 40, nil)

	if unsafe.Sizeof(uintptr(0)) > 4 {
		for name, value := range map[string]*gui.Window{"root": root, "parent": parent, "window": win} {
			if ptr := uintptr(unsafe.Pointer(value)); ptr <= uintptr(^uint32(0)) {
				t.Fatalf("%s pointer = %#x, want an address above 4 GiB", name, ptr)
			}
		}
	}

	if got := nox_xxx_wnd_46C2A0(nil); got != 1 {
		t.Fatalf("nil result = %d, want 1", got)
	}
	if got := nox_xxx_wnd_46C2A0(win); got != 0 {
		t.Fatalf("visible chain result = %d, want 0", got)
	}

	root.Flags |= gui.StatusHidden
	if got := nox_xxx_wnd_46C2A0(win); got != 1 {
		t.Fatalf("hidden root result = %d, want 1", got)
	}
	root.Flags &^= gui.StatusHidden
	parent.Flags |= gui.StatusHidden
	if got := nox_xxx_wnd_46C2A0(win); got != 1 {
		t.Fatalf("hidden parent result = %d, want 1", got)
	}
	parent.Flags &^= gui.StatusHidden
	win.Flags |= gui.StatusHidden
	if got := nox_xxx_wnd_46C2A0(win); got != 1 {
		t.Fatalf("hidden self result = %d, want 1", got)
	}

	win.Flags = gui.StatusDestroyed
	if got := nox_xxx_wnd_46C2A0(win); got != 0 {
		t.Fatalf("upper-byte-only flags result = %d, want 0", got)
	}
	win.Flags = gui.StatusEnabled
}
