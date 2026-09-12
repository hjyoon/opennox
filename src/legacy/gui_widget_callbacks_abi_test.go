package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func requireNativeWindowAddress(t *testing.T, win *gui.Window) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(win.C()) <= math.MaxUint32 {
		t.Fatalf("test window address = %p, want a pointer above the PE32 range", win.C())
	}
}

func TestEditEventCallbackPreservesNativeWindowPointer(t *testing.T) {
	g := gui.New(nil)
	defer g.DestroyAll()

	parent := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 640, 480, nil)
	draw := gui.WindowData{Window: parent, Style: gui.StyleEntryField}
	win := gui.NewEntryFieldRaw(g, parent, gui.StatusEnabled, 0, 0, 120, 20, &draw, &gui.EntryFieldData{Field_1040: 32})
	if win == nil {
		t.Fatal("NewEntryFieldRaw returned nil")
	}
	requireNativeWindowAddress(t, win)

	// This is the exact event and packed cursor position from the reported
	// crash. The PE32 implementation faulted while reading window offset 44.
	if got := editEventCallbackC(win, 17, 0x005c0210, 0); got != 1 {
		t.Fatalf("edit callback response = %d, want 1", got)
	}
}

func TestListBoxCallbacksPreserveNativePointers(t *testing.T) {
	for _, multi := range []bool{false, true} {
		t.Run(map[bool]string{false: "single", true: "multi"}[multi], func(t *testing.T) {
			g := gui.New(nil)
			defer g.DestroyAll()

			parent := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 640, 480, nil)
			draw := gui.WindowData{Window: parent, Style: gui.StyleScrollListBox | gui.StyleMouseTrack}
			opts := gui.ScrollListBoxData{Count: 4, Line_height: 10}
			if multi {
				opts.Field_4 = 1
			}
			win := gui.NewScrollListBoxRaw(g, parent, gui.StatusEnabled, 10, 20, 120, 60, &draw, &opts)
			if win == nil {
				t.Fatal("NewScrollListBoxRaw returned nil")
			}
			requireNativeWindowAddress(t, win)

			text := alloc.InternCString16("alpha")
			if got := listBoxPreCallbackC(win, 0x400d, uintptr(unsafe.Pointer(text)), ^uintptr(0)); got != 1 {
				t.Fatalf("add-line callback response = %d, want 1", got)
			}
			gotText := listBoxPreCallbackC(win, 0x4016, 0, 0)
			if gotText == 0 || alloc.GoString16((*uint16)(unsafe.Pointer(gotText))) != "alpha" {
				t.Fatalf("item text pointer = %#x, want alpha", gotText)
			}
			if unsafe.Sizeof(uintptr(0)) > 4 && gotText <= math.MaxUint32 {
				t.Fatalf("item text pointer = %#x, want a native-width address", gotText)
			}
			if got := listBoxEventCallbackC(win, multi, 17, 0x005c0210, 0); got != 1 {
				t.Fatalf("list-box callback response = %d, want 1", got)
			}

			// The legacy initializer used to reinstall PE32 draw and event
			// callbacks. It must retain the native Go callbacks instead.
			listBoxInitCallbackC(win)
			win.Draw()
		})
	}
}

func TestInputConfigCallbacksPreserveNativeWindowPointer(t *testing.T) {
	for _, shell := range []bool{false, true} {
		t.Run(map[bool]string{false: "in_game", true: "shell"}[shell], func(t *testing.T) {
			g := gui.New(nil)
			defer g.DestroyAll()

			parent := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 640, 480, nil)
			draw := gui.WindowData{Window: parent, Style: gui.StyleScrollListBox | gui.StyleMouseTrack}
			win := gui.NewScrollListBoxRaw(
				g, parent, gui.StatusEnabled, 10, 20, 120, 60, &draw,
				&gui.ScrollListBoxData{Count: 4, Line_height: 10},
			)
			if win == nil {
				t.Fatal("NewScrollListBoxRaw returned nil")
			}
			requireNativeWindowAddress(t, win)

			// Event 17 makes the list callback inspect the native Window, while
			// the control callback immediately reads its native widget data.
			if got := inputConfigListCallbackC(win, shell, 17, 0x005c0210, 0); got != 1 {
				t.Fatalf("list callback response = %d, want 1", got)
			}
			if got := inputConfigControlCallbackC(win, shell, 17, 0x005c0210, 0); got != 0 {
				t.Fatalf("control callback response = %d, want 0", got)
			}
		})
	}
}
