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
	win.Func94(AsWindowEvent(0x400B, 10, 10))
	if d := sliderData(win); d.Min != 10 || d.Max != 10 || d.Field3 != 10 {
		t.Fatalf("zero slider range/current = (%d, %d, %d), want (10, 10, 10)", d.Min, d.Max, d.Field3)
	}
	if got := win.Field100().Offs().Y; got != 0 {
		t.Fatalf("zero-range vertical thumb Y = %d, want 0", got)
	}
	win.Destroy()
	g.FreeDestroyed()
}

func TestVerticalImageSliderDrawsOnlyMovingThumb(t *testing.T) {
	g := New(nil)
	defer g.alloc.Free()

	verticalDraw := WindowData{Style: StyleVertSlider}
	vertical := NewSliderRaw(g, nil, StatusEnabled|StatusImage, 0, 0, 10, 50, &verticalDraw, &SliderData{})
	if vertical == nil {
		t.Fatal("vertical image slider was not created")
	}
	if sliderDrawsImageTrack(vertical) {
		t.Fatal("vertical image slider would draw a fixed duplicate thumb")
	}

	horizontalDraw := WindowData{Style: StyleHorizSlider}
	horizontal := NewSliderRaw(g, nil, StatusEnabled|StatusImage, 0, 0, 50, 10, &horizontalDraw, &SliderData{})
	if horizontal == nil {
		t.Fatal("horizontal image slider was not created")
	}
	if !sliderDrawsImageTrack(horizontal) {
		t.Fatal("horizontal image slider lost its background track")
	}

	vertical.Destroy()
	horizontal.Destroy()
	g.FreeDestroyed()
}

func TestSliderDragUsesMovedThumbAndActualSize(t *testing.T) {
	g := New(nil)
	defer g.alloc.Free()

	var changed uint32
	parent := g.NewWindowRaw(nil, StatusEnabled, 0, 0, 300, 200, func(_ *Window, e WindowEvent) WindowEventResp {
		if e.EventCode() == 0x4009 {
			_, value := e.EventArgsC()
			changed = uint32(value)
		}
		return RawEventResp(1)
	})
	draw := WindowData{Window: parent, Style: StyleVertSlider | StyleMouseTrack}
	slider := NewSliderRaw(g, parent, StatusEnabled, 20, 10, 16, 110, &draw, &SliderData{Min: 0, Max: 100})
	if slider == nil || slider.Field100() == nil {
		t.Fatal("native slider or thumb was not created")
	}
	thumb := slider.Field100()
	thumb.SizeVal = image.Pt(16, 20)
	thumb.SetEnd(thumb.Offs().Add(thumb.Size()))
	slider.Func94(AsWindowEvent(0x400B, 0, 100))

	if got := sliderTrackLength(slider); got != 90 {
		t.Fatalf("track length = %d, want 90", got)
	}
	if got := thumb.Offs().Y; got != 90 {
		t.Fatalf("minimum-value thumb Y = %d, want 90", got)
	}

	slider.Func94(AsWindowEvent(0x400A, 50, 0))
	if got := thumb.Offs().Y; got != 45 {
		t.Fatalf("midpoint thumb Y = %d, want 45", got)
	}

	// The input loop moves draggable windows before dispatching PRESSED. The
	// slider must read that physical position instead of recentering the thumb
	// around the cursor (which can be anywhere inside the 20-pixel thumb).
	thumb.SetPos(image.Pt(7, 30))
	cursorNearTop := slider.GlobalPos().Add(image.Pt(2, 31))
	thumb.Func93(&WindowMouseState{State: input.NOX_MOUSE_LEFT_PRESSED, Pos: cursorNearTop})
	if got := sliderData(slider).Field3; got != 67 {
		t.Fatalf("value after moved-thumb drag = %d, want 67", got)
	}
	if changed != 67 {
		t.Fatalf("notified value after moved-thumb drag = %d, want 67", changed)
	}
	if got := thumb.Offs(); got != image.Pt(0, 30) {
		t.Fatalf("thumb position after drag = %v, want (0,30)", got)
	}

	// A release may be the first event observed at the final mouse position.
	// Synchronize from the already-moved thumb before sending completion.
	thumb.Func93(&WindowMouseState{State: input.NOX_MOUSE_LEFT_DOWN, Pos: cursorNearTop})
	thumb.SetPos(image.Pt(0, sliderTrackLength(slider)))
	thumb.Func93(&WindowMouseState{State: input.NOX_MOUSE_LEFT_UP, Pos: cursorNearTop})
	if got := sliderData(slider).Field3; got != 0 {
		t.Fatalf("value after release at bottom = %d, want 0", got)
	}

	parent.Destroy()
	g.FreeDestroyed()
}

func TestScrollListBoxSliderDragKeepsThumbAtBottom(t *testing.T) {
	g := New(nil)
	defer g.alloc.Free()

	parent := g.NewWindowRaw(nil, StatusEnabled, 0, 0, 300, 200, nil)
	draw := WindowData{Window: parent, Style: StyleScrollListBox | StyleMouseTrack}
	win := NewScrollListBoxRaw(g, parent, StatusEnabled, 10, 20, 120, 60, &draw, &ScrollListBoxData{
		Count:       10,
		Line_height: 10,
		Field_3:     1,
	})
	if win == nil {
		t.Fatal("NewScrollListBoxRaw returned nil")
	}
	for i := 0; i < 10; i++ {
		if !scrollListBoxAddLine(win, "line", -1) {
			t.Fatalf("failed to add line %d", i)
		}
	}

	d := scrollListBoxData(win)
	slider := scrollListBoxWindow(d.Field_9)
	if slider == nil || slider.Field100() == nil {
		t.Fatal("listbox slider or thumb was not created")
	}
	sd := sliderData(slider)
	if sd.Max != 53 {
		t.Fatalf("slider maximum = %d, want 53", sd.Max)
	}

	// Exercise the same path as holding and dragging the thumb: the input loop
	// first moves the draggable child, then the button forwards PRESSED to the
	// slider. Keep the cursor off-center to catch unwanted drag jumps.
	thumb := slider.Field100()
	thumb.SetPos(image.Pt(0, sliderTrackLength(slider)))
	nearThumbTop := slider.GlobalPos().Add(image.Pt(slider.Size().X/2, sliderTrackLength(slider)+1))
	thumb.Func93(&WindowMouseState{State: input.NOX_MOUSE_LEFT_PRESSED, Pos: nearThumbTop})

	if got := int(d.Field_13_1); got != 51 {
		t.Fatalf("listbox bottom offset = %d, want 51", got)
	}
	if got := sd.Field3; got != sd.Min {
		t.Fatalf("slider value after bottom drag = %d, want minimum %d", got, sd.Min)
	}
	if got, want := slider.Field100().Offs().Y, sliderTrackLength(slider); got != want {
		t.Fatalf("thumb Y after bottom drag = %d, want %d", got, want)
	}

	win.Destroy()
	g.FreeDestroyed()
}

func TestScrollListBoxResizeRelayoutsAndKeepsControlsWorking(t *testing.T) {
	g := New(nil)
	defer g.alloc.Free()

	parent := g.NewWindowRaw(nil, StatusEnabled, 0, 0, 300, 200, nil)
	draw := WindowData{Window: parent, Style: StyleScrollListBox | StyleMouseTrack}
	win := NewScrollListBoxRaw(g, parent, StatusEnabled, 10, 20, 120, 60, &draw, &ScrollListBoxData{
		Count:       10,
		Line_height: 10,
		Field_3:     1,
	})
	if win == nil {
		t.Fatal("NewScrollListBoxRaw returned nil")
	}
	for i := 0; i < 10; i++ {
		if !scrollListBoxAddLine(win, "line", -1) {
			t.Fatalf("failed to add line %d", i)
		}
	}

	if got := win.SetSize(image.Pt(150, 90)); got != 0 {
		t.Fatalf("SetSize returned %d, want 0", got)
	}
	d := scrollListBoxData(win)
	up := scrollListBoxWindow(d.Field_7)
	down := scrollListBoxWindow(d.Field_8)
	slider := scrollListBoxWindow(d.Field_9)
	if up == nil || down == nil || slider == nil || slider.Field100() == nil {
		t.Fatal("resized listbox controls are missing")
	}
	if got := up.Offs(); got != image.Pt(140, 0) {
		t.Fatalf("up button position = %v, want (140,0)", got)
	}
	if got := down.Offs(); got != image.Pt(140, 80) {
		t.Fatalf("down button position = %v, want (140,80)", got)
	}
	if got := slider.Offs(); got != image.Pt(140, 10) {
		t.Fatalf("slider position = %v, want (140,10)", got)
	}
	if got := slider.Size(); got != image.Pt(10, 70) {
		t.Fatalf("slider size = %v, want (10,70)", got)
	}
	if got := d.Field_13_0; got != 90 {
		t.Fatalf("viewport height = %d, want 90", got)
	}
	sd := sliderData(slider)
	if sd == nil || sd.Max != 23 {
		t.Fatalf("resized slider range = %+v, want maximum 23", sd)
	}
	if got := sliderTrackLength(slider); got != 60 {
		t.Fatalf("resized slider track = %d, want 60", got)
	}

	// Exercise the actual button event path after the relayout.
	down.Func93(&WindowMouseState{State: input.NOX_MOUSE_LEFT_DOWN})
	down.Func93(&WindowMouseState{State: input.NOX_MOUSE_LEFT_UP})
	if got := d.Field_13_1; got != 12 {
		t.Fatalf("offset after down button = %d, want 12", got)
	}
	up.Func93(&WindowMouseState{State: input.NOX_MOUSE_LEFT_DOWN})
	up.Func93(&WindowMouseState{State: input.NOX_MOUSE_LEFT_UP})
	if got := d.Field_13_1; got != 0 {
		t.Fatalf("offset after up button = %d, want 0", got)
	}

	// Then drag the resized track to its lower endpoint. The list offset is
	// clamped to content height while the thumb remains at the endpoint.
	thumb := slider.Field100()
	thumb.SetPos(image.Pt(0, sliderTrackLength(slider)))
	nearThumbTop := slider.GlobalPos().Add(image.Pt(slider.Size().X/2, sliderTrackLength(slider)+1))
	thumb.Func93(&WindowMouseState{State: input.NOX_MOUSE_LEFT_PRESSED, Pos: nearThumbTop})
	if got := d.Field_13_1; got != 21 {
		t.Fatalf("bottom offset after resize = %d, want 21", got)
	}
	if got := thumb.Offs().Y; got != sliderTrackLength(slider) {
		t.Fatalf("thumb Y after resized drag = %d, want %d", got, sliderTrackLength(slider))
	}

	parent.Destroy()
	g.FreeDestroyed()
}

func TestScrollListBoxMouseWheel(t *testing.T) {
	g := New(nil)
	defer g.alloc.Free()

	parent := g.NewWindowRaw(nil, StatusEnabled, 0, 0, 300, 200, nil)
	newList := func(y int, multiple bool) *Window {
		draw := WindowData{Window: parent, Style: StyleScrollListBox | StyleMouseTrack}
		multi := uint32(0)
		if multiple {
			multi = 1
		}
		win := NewScrollListBoxRaw(g, parent, StatusEnabled, 10, y, 120, 30, &draw, &ScrollListBoxData{
			Count:       6,
			Line_height: 10,
			Field_3:     1,
			Field_4:     multi,
		})
		if win == nil {
			t.Fatal("NewScrollListBoxRaw returned nil")
		}
		for i := 0; i < 6; i++ {
			if !scrollListBoxAddLine(win, "line", -1) {
				t.Fatalf("failed to add line %d", i)
			}
		}
		return win
	}

	single := newList(0, false)
	if !EventRespBool(single.Func93(&WindowMouseState{State: input.MouseStateCode(20)})) {
		t.Fatal("single-select wheel-down event was not handled")
	}
	if got := scrollListBoxSelection(single); got != 0 {
		t.Fatalf("selection after first wheel-down = %d, want 0", got)
	}
	single.Func93(&WindowMouseState{State: input.MouseStateCode(20)})
	if got := scrollListBoxSelection(single); got != 1 {
		t.Fatalf("selection after second wheel-down = %d, want 1", got)
	}
	single.Func93(&WindowMouseState{State: input.MouseStateCode(19)})
	if got := scrollListBoxSelection(single); got != 0 {
		t.Fatalf("selection after wheel-up = %d, want 0", got)
	}

	multiple := newList(40, true)
	md := scrollListBoxData(multiple)
	multiple.Func93(&WindowMouseState{State: input.MouseStateCode(20)})
	if got := md.Field_13_1; got == 0 {
		t.Fatal("multi-select wheel-down did not scroll the viewport")
	}
	multiple.Func93(&WindowMouseState{State: input.MouseStateCode(19)})
	if got := md.Field_13_1; got != 0 {
		t.Fatalf("multi-select wheel-up offset = %d, want 0", got)
	}

	parent.Destroy()
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
