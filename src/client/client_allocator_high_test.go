package client

import (
	"image"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestClientExplicitHighOwnerIsolation(t *testing.T) {
	a, b := server.New(nil, nil, strman.New()), server.NewHigh(nil, nil, strman.New())
	t.Cleanup(a.Close)
	t.Cleanup(b.Close)
	ordinary, high := NewClient(nil, nil, a), NewClientHigh(nil, nil, b)
	for _, cl := range []*Client{ordinary, high} {
		t.Cleanup(func() {
			cl.GUI.DestroyAll()
			alloc.Free(cl.vp)
			cl.Close()
		})
	}
	if ordinary.vp == high.vp || *ordinary.vp != *high.vp {
		t.Fatal("viewport constructors lost independent zeroed storage")
	}
	win := high.GUI.NewWindowRaw(nil, gui.StatusEnabled, 1, 2, 3, 4, nil)
	if win == nil {
		t.Fatal("explicit client GUI did not allocate a window")
	}
	for _, ptr := range []unsafe.Pointer{high.vp.C(), win.C()} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
			t.Fatalf("explicit client pointer %p lacks high native address bits", ptr)
		}
	}
	ordinary.vp.Screen = image.Rect(1, 2, 3, 4)
	high.vp.Screen = image.Rect(5, 6, 7, 8)
	if ordinary.vp.Screen != image.Rect(1, 2, 3, 4) || high.vp.Screen != image.Rect(5, 6, 7, 8) {
		t.Fatal("viewport mutations crossed client ownership")
	}
}
