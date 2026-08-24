package gui

import (
	"image"
	"math"
	"unsafe"

	"github.com/opennox/libs/client/keybind"

	"github.com/opennox/opennox/v1/client/input"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type SliderData struct {
	Min    uint32
	Max    uint32
	Field2 uint32
	Field3 uint32
}

func (d *SliderData) CWidgetData() unsafe.Pointer {
	return unsafe.Pointer(d)
}

// NewSliderRaw is the native-width replacement for nox_gui_newSlider_4B4EE0.
// In particular, it keeps the parent, WindowData and child thumb pointers out
// of the original PE32 int ABI.
func NewSliderRaw(g *GUI, parent *Window, status StatusFlags, px, py, w, h int, draw *WindowData, opts *SliderData) *Window {
	if g == nil || draw == nil || opts == nil || (!draw.Style.IsVertSlider() && !draw.Style.IsHorizSlider()) {
		return nil
	}
	win := g.NewWindowRaw(parent, status|StatusTabStop, px, py, w, h, sliderProcPre)
	if win == nil {
		return nil
	}
	win.SetAllFuncs(sliderProc, sliderDraw, nil)
	if draw.Window == nil {
		draw.Window = win
	}
	win.CopyDrawData(draw)
	d, _ := alloc.New(SliderData{})
	*d = *opts
	if d.Max == d.Min {
		d.Max = d.Min + 1
	}
	if d.Field3 < d.Min || d.Field3 > d.Max {
		d.Field3 = d.Min
	}
	win.WidgetData = unsafe.Pointer(d)
	sliderRecalculate(win)

	thumbDraw := *draw
	thumbDraw.Window = win
	thumbDraw.Style = StylePushButton | (draw.Style & StyleMouseTrack)
	thumbDraw.SetText("")
	thumbStatus := (status &^ StatusHidden) | StatusEnabled | StatusDraggable | StatusNoFocus
	thumbW, thumbH := 10, h
	if draw.Style.IsVertSlider() {
		thumbW, thumbH = w, 10
	}
	thumb := NewButtonRaw(g, win, thumbStatus, 0, 0, thumbW, thumbH, &thumbDraw)
	if thumb == nil {
		return nil
	}
	sliderPositionThumb(win)
	return win
}

func sliderData(win *Window) *SliderData {
	if win == nil || win.WidgetData == nil {
		return nil
	}
	return (*SliderData)(win.WidgetData)
}

func sliderVertical(win *Window) bool {
	return win != nil && win.DrawData().Style.IsVertSlider()
}

func sliderTrackLength(win *Window) int {
	if sliderVertical(win) {
		return max(win.Size().Y-10, 0)
	}
	return max(win.Size().X-10, 0)
}

func sliderRecalculate(win *Window) {
	d := sliderData(win)
	if d == nil {
		return
	}
	if d.Max <= d.Min {
		d.Max = d.Min + 1
	}
	step := float32(sliderTrackLength(win)) / float32(d.Max-d.Min)
	d.Field2 = math.Float32bits(step)
	d.Field3 = min(max(d.Field3, d.Min), d.Max)
	sliderPositionThumb(win)
}

func sliderPositionThumb(win *Window) {
	d := sliderData(win)
	if d == nil {
		return
	}
	thumb := win.Field100()
	if thumb == nil {
		return
	}
	step := math.Float32frombits(d.Field2)
	if sliderVertical(win) {
		y := int(math.Round(float64(float32(d.Max-d.Field3) * step)))
		thumb.SetPos(image.Pt(0, min(max(y, 0), sliderTrackLength(win))))
	} else {
		x := int(math.Round(float64(float32(d.Field3-d.Min) * step)))
		thumb.SetPos(image.Pt(min(max(x, 0), sliderTrackLength(win)), 0))
	}
}

func sliderNotify(win *Window, code int, a1, a2 uintptr) {
	if parent := win.DrawData().Window; parent != nil {
		parent.Func94(AsWindowEvent(code, a1, a2))
	}
}

func sliderSetValue(win *Window, value uint32, notify bool) {
	d := sliderData(win)
	if d == nil {
		return
	}
	d.Field3 = min(max(value, d.Min), d.Max)
	sliderPositionThumb(win)
	if notify {
		sliderNotify(win, 0x4009, uintptr(unsafe.Pointer(win)), uintptr(d.Field3))
	}
}

func sliderValueAt(win *Window, pos image.Point) uint32 {
	d := sliderData(win)
	if d == nil {
		return 0
	}
	global := win.GlobalPos()
	coord := pos.X - global.X - 5
	if sliderVertical(win) {
		coord = pos.Y - global.Y - 5
	}
	coord = min(max(coord, 0), sliderTrackLength(win))
	step := math.Float32frombits(d.Field2)
	if step <= 0 {
		return d.Min
	}
	delta := uint32(math.Round(float64(float32(coord) / step)))
	if sliderVertical(win) {
		if delta > d.Max-d.Min {
			delta = d.Max - d.Min
		}
		return d.Max - delta
	}
	return min(d.Min+delta, d.Max)
}

func sliderProcPre(win *Window, e WindowEvent) WindowEventResp {
	d := sliderData(win)
	if d == nil {
		return RawEventResp(0)
	}
	a1, a2 := e.EventArgsC()
	switch e := e.(type) {
	case WindowDestroy:
		alloc.FreePtr(win.WidgetData)
		win.WidgetData = nil
		return RawEventResp(0)
	case WindowFocus:
		win.DrawData().Field0Set(0x2, bool(e))
		sliderNotify(win, 0x4003, a1, uintptr(win.ID()))
		return RawEventResp(1)
	}
	switch e.EventCode() {
	case 0x4000:
		sx := uint16(a2)
		sy := uint16(a2 >> 16)
		sliderSetValue(win, sliderValueAt(win, image.Pt(int(sx), int(sy))), true)
	case 0x4007:
		sliderNotify(win, 0x400C, uintptr(unsafe.Pointer(win)), uintptr(d.Field3))
	case 0x400A:
		sliderSetValue(win, uint32(a1), false)
	case 0x400B:
		d.Min, d.Max = uint32(a1), uint32(a2)
		d.Field3 = d.Min
		sliderRecalculate(win)
	}
	return RawEventResp(0)
}

func sliderProc(win *Window, e WindowEvent) WindowEventResp {
	switch e := e.(type) {
	case *WindowMouseState:
		switch e.State {
		case input.NOX_MOUSE_LEFT_DOWN, input.NOX_MOUSE_LEFT_DRAG_END, input.NOX_MOUSE_LEFT_UP:
			sliderSetValue(win, sliderValueAt(win, e.Pos), true)
			return RawEventResp(1)
		case input.NOX_MOUSE_LEFT_PRESSED:
			sliderNotify(win, 0x4000, uintptr(unsafe.Pointer(win)), 0)
			return RawEventResp(1)
		}
	case *WindowMouseUnk:
		switch e.Event {
		case 17:
			win.DrawData().Field0 |= 0x2
			sliderNotify(win, 0x4005, uintptr(unsafe.Pointer(win)), 0)
			win.Focus()
			return RawEventResp(1)
		case 18:
			win.DrawData().Field0 &^= 0x2
			sliderNotify(win, 0x4006, uintptr(unsafe.Pointer(win)), 0)
			return RawEventResp(1)
		}
	case WindowKeyPress:
		if !e.Pressed {
			return RawEventResp(1)
		}
		d := sliderData(win)
		if d == nil {
			return RawEventResp(0)
		}
		switch e.Key {
		case keybind.KeyUp, keybind.KeyRight:
			if d.Field3 < d.Max {
				sliderSetValue(win, d.Field3+1, true)
			}
			return RawEventResp(1)
		case keybind.KeyDown, keybind.KeyLeft:
			if d.Field3 > d.Min {
				sliderSetValue(win, d.Field3-1, true)
			}
			return RawEventResp(1)
		case keybind.KeyTab, keybind.KeyEnter, keybind.KeySpace:
			return RawEventResp(1)
		}
	}
	return RawEventResp(0)
}

func sliderDraw(win *Window, draw *WindowData) int {
	r := win.GUI().Render()
	if r == nil {
		return 1
	}
	pos := win.GlobalPos()
	w, h := win.Size().X, win.Size().Y
	if win.Flags.Has(StatusImage) {
		img := draw.BackgroundImage()
		if !win.Flags.IsEnabled() {
			img = draw.DisabledImage()
		}
		if img != nil {
			r.DrawImage16(img, pos.Add(draw.ImagePoint()))
		}
		return 1
	}
	bg := draw.BackgroundColor()
	if !win.Flags.IsEnabled() {
		bg = draw.DisabledColor()
	}
	if _, _, _, a := bg.RGBA(); a != 0 {
		r.DrawRectFilledOpaque(pos.X+1, pos.Y+1, max(w-2, 0), max(h-2, 0), bg)
	}
	track := draw.EnabledColor()
	if draw.Field0&0x2 != 0 {
		track = draw.HighlightColor()
	}
	if _, _, _, a := track.RGBA(); a != 0 {
		if sliderVertical(win) {
			x := pos.X + w/2
			r.DrawRectFilledOpaque(x-1, pos.Y+4, 3, max(h-8, 0), track)
		} else {
			y := pos.Y + h/2
			r.DrawRectFilledOpaque(pos.X, y-1, w, 3, track)
		}
	}
	return 1
}
