package gui

import (
	"image"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client/input"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestScrollListBoxNativeEvents(t *testing.T) {
	g := New(nil)
	defer g.alloc.Free()

	var selected int
	parent := g.NewWindowRaw(nil, StatusEnabled, 0, 0, 300, 200, func(_ *Window, e WindowEvent) WindowEventResp {
		if e.EventCode() == 0x4010 {
			_, a2 := e.EventArgsC()
			selected = int(a2)
		}
		return RawEventResp(1)
	})
	draw := WindowData{Window: parent, Style: StyleScrollListBox | StyleMouseTrack}
	win := NewScrollListBoxRaw(g, parent, StatusEnabled, 10, 20, 120, 60, &draw, &ScrollListBoxData{
		Count:       4,
		Line_height: 10,
	})
	if win == nil {
		t.Fatal("NewScrollListBoxRaw returned nil")
	}
	if !scrollListBoxAddLine(win, "alpha", -1) || !scrollListBoxAddLine(win, "beta", -1) {
		t.Fatal("failed to add native listbox lines")
	}
	d := scrollListBoxData(win)
	if d.Field_11_0 != 2 || d.Field_10 != 22 {
		t.Fatalf("listbox counts = (%d, %d), want (2, 22)", d.Field_11_0, d.Field_10)
	}

	scrollListBoxProc(win, &WindowMouseState{State: input.NOX_MOUSE_LEFT_UP, Pos: image.Pt(15, 25)})
	if selected != 0 || scrollListBoxSelection(win) != 0 {
		t.Fatalf("selection = (%d, %d), want (0, 0)", selected, scrollListBoxSelection(win))
	}
	win.Func94(AsWindowEvent(0x4013, 1, 0))
	if got := EventRespInt(win.Func94(AsWindowEvent(0x4014, 0, 0))); got != 1 {
		t.Fatalf("programmatic selection = %d, want 1", got)
	}
	if got := EventRespPtr(win.Func94(AsWindowEvent(0x4016, 1, 0))); got == nil {
		t.Fatal("0x4016 returned a nil item string")
	}
	replacement := alloc.InternCString16("gamma")
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(replacement)) <= math.MaxUint32 {
		t.Fatalf("replacement string pointer = %p, want native address above 4 GiB", replacement)
	}
	win.Func94(AsWindowEvent(0x4017, uintptr(unsafe.Pointer(replacement)), 1))
	if got := alloc.GoString16(&scrollListBoxItems(d)[1].Text[0]); got != "gamma" {
		t.Fatalf("updated item text = %q, want gamma", got)
	}
	win.Func94(AsWindowEvent(0x400E, 0, 0))
	if d.Field_11_0 != 1 {
		t.Fatalf("line count after delete = %d, want 1", d.Field_11_0)
	}
	win.Func94(AsWindowEvent(0x400F, 0, 0))
	if d.Field_11_0 != 0 || d.Field_10 != 0 {
		t.Fatalf("listbox was not cleared: count=%d height=%d", d.Field_11_0, d.Field_10)
	}
	win.Destroy()
	g.FreeDestroyed()
}

func TestSliderNativeRangeAndThumb(t *testing.T) {
	g := New(nil)
	defer g.alloc.Free()

	draw := WindowData{Style: StyleVertSlider}
	win := NewSliderRaw(g, nil, StatusEnabled, 0, 0, 16, 110, &draw, &SliderData{Min: 0, Max: 100})
	if win == nil || win.Field100() == nil {
		t.Fatal("native slider or thumb was not created")
	}
	win.Func94(AsWindowEvent(0x400A, 75, 0))
	if got := sliderData(win).Field3; got != 75 {
		t.Fatalf("slider value = %d, want 75", got)
	}
	if got := win.Field100().Offs().Y; got != 25 {
		t.Fatalf("vertical thumb Y = %d, want 25", got)
	}
	win.Func94(AsWindowEvent(0x400B, 10, 30))
	if d := sliderData(win); d.Min != 10 || d.Max != 30 || d.Field3 != 10 {
		t.Fatalf("slider range/current = (%d, %d, %d), want (10, 30, 10)", d.Min, d.Max, d.Field3)
	}
	win.Destroy()
	g.FreeDestroyed()
}

func TestEntryFieldNativeSetAndGet(t *testing.T) {
	g := New(nil)
	defer g.alloc.Free()

	parent := g.NewWindowRaw(nil, StatusEnabled, 0, 0, 300, 200, nil)
	draw := WindowData{Window: parent, Style: StyleEntryField}
	win := NewEntryFieldRaw(g, parent, StatusEnabled, 10, 20, 120, 20, &draw, &EntryFieldData{Field_1040: 8})
	if win == nil {
		t.Fatal("NewEntryFieldRaw returned nil")
	}
	str := alloc.InternCString16("abcdefghi")
	win.Func94(AsWindowEvent(0x401E, uintptr(unsafe.Pointer(str)), 0))
	d := entryFieldData(win)
	if got := alloc.GoString16(&d.Text[0]); got != "abcdefg" {
		t.Fatalf("entry text = %q, want %q", got, "abcdefg")
	}
	if got := EventRespPtr(win.Func94(AsWindowEvent(0x401D, 0, 0))); got != win.WidgetData {
		t.Fatalf("entry data pointer = %p, want %p", got, win.WidgetData)
	}
	win.Destroy()
	g.FreeDestroyed()
}
