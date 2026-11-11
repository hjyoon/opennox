package legacy

import (
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestGUIExplicitHighWindowOwnership(t *testing.T) {
	ordinary, high := gui.New(nil), gui.NewHigh(nil)
	defer ordinary.DestroyAll()
	defer high.DestroyAll()
	w := high.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 10, 20, nil)
	other := ordinary.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 30, 40, nil)
	if runtime.GOOS == "windows" && unsafe.Sizeof(uintptr(0)) == 8 && uintptr(w.C()) <= uintptr(^uint32(0)) {
		t.Fatal("explicit GUI fixture did not exercise high window address bits")
	}
	w.SetID(77)
	other.SetID(99)
	cleanup := 0
	var duringCleanup *gui.Window
	w.AddCleanup(func() {
		cleanup++
		if !high.OwnsWindow(w) || w.GUI() != high || w.SizeVal.X != 10 || w.SizeVal.Y != 20 {
			t.Error("window storage was recycled before its cleanup completed")
		}
		duringCleanup = high.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 5, 6, nil)
		if duringCleanup == w {
			t.Error("reentrant allocation recycled a window whose cleanup was still running")
		}
	})
	runtime.GC()
	if w.GUI() != high || other.GUI() != ordinary {
		t.Fatal("window allocator selection leaked between GUI owners")
	}
	w.Destroy()
	high.FreeDestroyed()
	if cleanup != 1 || high.OwnsWindow(w) || !ordinary.OwnsWindow(other) || other.ID() != 99 {
		t.Fatal("high window release changed cleanup ownership or the ordinary GUI")
	}
	reused := high.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 50, 60, nil)
	if reused != w || reused.ID() != 0 || reused.GUI() != high {
		t.Fatal("high window slot did not reuse cleared storage with fresh ownership")
	}
	draw := gui.WindowData{Window: reused, Style: gui.StyleScrollListBox}
	opts := gui.ScrollListBoxData{Count: 2, Line_height: 10, Field_4: 1}
	list := gui.NewScrollListBoxRaw(high, reused, gui.StatusEnabled, 0, 0, 100, 20, &draw, &opts)
	data := (*gui.ScrollListBoxData)(list.WidgetData)
	// This raw callback transports uintptr, so its string input must be
	// foreign storage rather than a Go-heap pointer with erased provenance.
	text, freeText := alloc.MakeHigh([]uint16{'x', 0}, 2)
	defer freeText()
	if !gui.EventRespBool(list.Func94(gui.AsWindowEvent(0x400d, uintptr(unsafe.Pointer(&text[0])), ^uintptr(0)))) {
		t.Fatal("high widget did not copy the declared input text")
	}
	runtime.KeepAlive(text)
	output := gui.EventRespPtr(list.Func94(gui.AsWindowEvent(0x4016, 0, 0)))
	if output != unsafe.Pointer(&data.Items.Text[0]) || *(*uint16)(output) != 'x' {
		t.Fatal("high widget did not return its owned inline text")
	}
	if runtime.GOOS == "windows" && unsafe.Sizeof(uintptr(0)) == 8 {
		for _, p := range []unsafe.Pointer{list.WidgetData, unsafe.Pointer(data.Items), unsafe.Pointer(data.Field_12), output} {
			if uintptr(p) <= uintptr(^uint32(0)) {
				t.Fatal("high GUI widget storage did not inherit its owner allocation policy")
			}
		}
	}
	high.DestroyAll()
	if high.OwnsWindow(reused) || high.OwnsWindow(list) || high.OwnsWindow(duringCleanup) || cleanup != 1 {
		t.Fatal("high GUI destruction retained ownership or repeated an old cleanup")
	}
}
