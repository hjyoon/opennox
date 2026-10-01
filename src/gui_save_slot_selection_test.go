package opennox

import (
	"fmt"
	"image"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
)

func TestSaveListSelectedSlotNative(t *testing.T) {
	g := gui.New(nil)
	t.Cleanup(g.DestroyAll)
	draw := gui.WindowData{Style: gui.StyleScrollListBox | gui.StyleMouseTrack}
	win := gui.NewScrollListBoxRaw(g, nil, gui.StatusEnabled, 0, 0, 200, 300, &draw,
		&gui.ScrollListBoxData{Count: NOX_SAVEGAME_XXX_MAX, Line_height: 17})
	if win == nil {
		t.Fatal("native save list constructor returned nil")
	}
	d := (*gui.ScrollListBoxData)(win.WidgetData)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if uintptr(unsafe.Pointer(win)) <= uintptr(^uint32(0)) || uintptr(win.WidgetData) <= uintptr(^uint32(0)) {
			t.Fatal("save list must exercise native pointers above 4 GiB")
		}
		if unsafe.Offsetof(d.Field_9) != 48 {
			t.Fatalf("LP64 slider offset=%d, want the former PE32 selection offset 48", unsafe.Offsetof(d.Field_9))
		}
	}
	for i := range NOX_SAVEGAME_XXX_MAX {
		win.Func94(&WindowEvent0x400d{Str: fmt.Sprintf("slot %d", i), Val: -1})
	}
	if ind, ok := saveListSelectedSlot(win, NOX_SAVEGAME_XXX_MAX); ind != -1 || ok {
		t.Fatalf("unselected list = (%d, %t), want (-1, false)", ind, ok)
	}
	for _, selected := range []int{0, 1, NOX_SAVEGAME_XXX_MAX - 1} {
		win.Func94(gui.AsWindowEvent(0x4013, uintptr(selected), 0))
		before := *d
		if ind, ok := saveListSelectedSlot(win, NOX_SAVEGAME_XXX_MAX); ind != selected || !ok {
			t.Fatalf("native selected slot = (%d, %t), want (%d, true)", ind, ok, selected)
		}
		if *d != before {
			t.Fatal("read-only slot query changed widget data")
		}
	}
	// A real unrelated native Window pointer in the LP64 slider field must
	// never become an array index or replace the list's selected slot.
	slider := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 10, 10, nil)
	d.Field_9 = unsafe.Pointer(slider)
	t.Cleanup(func() { d.Field_9 = nil })
	win.Func94(gui.AsWindowEvent(0x4013, 1, 0))
	if ind, ok := saveListSelectedSlot(win, NOX_SAVEGAME_XXX_MAX); ind != 1 || !ok {
		t.Fatalf("native slot with slider pointer = (%d, %t), want (1, true)", ind, ok)
	}
	if slider.Size() != image.Pt(10, 10) {
		t.Fatal("query changed the unrelated slider window")
	}
	if ind, ok := saveListSelectedSlot(win, 1); ind != 1 || ok {
		t.Fatalf("slot beyond save array = (%d, %t), want (1, false)", ind, ok)
	}
}

func TestSaveListSelectedSlotBounds(t *testing.T) {
	if ind, ok := saveListSelectedSlot(nil, NOX_SAVEGAME_XXX_MAX); ind != -1 || ok {
		t.Fatalf("nil list = (%d, %t), want (-1, false)", ind, ok)
	}
	g := gui.New(nil)
	t.Cleanup(g.DestroyAll)
	win := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 10, 10, nil)
	for _, value := range []int{-2, -1, 0, 1, NOX_SAVEGAME_XXX_MAX - 1, NOX_SAVEGAME_XXX_MAX, 1 << 30} {
		calls := 0
		win.SetFunc94(func(_ *gui.Window, ev gui.WindowEvent) gui.WindowEventResp {
			calls++
			a, b := ev.EventArgsC()
			if ev.EventCode() != 0x4014 || a != 0 || b != 0 {
				t.Fatal("slot selection used a write or a nonzero query argument")
			}
			return gui.RawEventResp(uintptr(uint32(int32(value))))
		})
		for _, count := range []int{-1, 0, 1, NOX_SAVEGAME_XXX_MAX} {
			before := calls
			ind, ok := saveListSelectedSlot(win, count)
			want := value >= 0 && value < count
			if ok != want || (count > 0 && ind != value) || (count <= 0 && calls != before) {
				t.Fatalf("value %d count %d returned (%d, %t), want valid=%t", value, count, ind, ok, want)
			}
		}
	}
	win.SetFunc94(nil)
	if _, ok := saveListSelectedSlot(win, NOX_SAVEGAME_XXX_MAX); ok {
		t.Fatal("nil native response was treated as autosave slot 0")
	}
	win.Destroy()
	if _, ok := saveListSelectedSlot(win, NOX_SAVEGAME_XXX_MAX); ok {
		t.Fatal("destroyed list was treated as autosave slot 0")
	}
}
