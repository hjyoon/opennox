package gui

import (
	"image"
	"unsafe"

	"github.com/opennox/libs/client/keybind"
	noxcolor "github.com/opennox/libs/color"

	"github.com/opennox/opennox/v1/client/input"
)

type RadioButtonData struct {
	Field0 uint32
}

func (d *RadioButtonData) CWidgetData() unsafe.Pointer {
	return unsafe.Pointer(d)
}

// RadioButtonProcPre is the native-width implementation of
// nox_xxx_wndRadioButtonProcPre_4A93C0.
func RadioButtonProcPre(win *Window, e WindowEvent) WindowEventResp {
	switch e := e.(type) {
	case *StaticTextSetText:
		win.DrawData().SetText(e.Str)
		return RawEventResp(0)
	case WindowFocus:
		if !e {
			win.DrawData().Field0 &^= 0x2
		}
		a1, _ := e.EventArgsC()
		radioButtonNotify(win, 0x4003, a1, uintptr(win.ID()))
		return RawEventResp(1)
	default:
		if e.EventCode() != 0x4008 {
			return RawEventResp(0)
		}
		a1, _ := e.EventArgsC()
		if win.DrawData().Field0&0x4 == 0 {
			if a1 == 1 {
				radioButtonNotify(win, 0x4007, uintptr(unsafe.Pointer(win)), 0)
			}
			radioButtonSelect(win)
		}
		return RawEventResp(0)
	}
}

// RadioButtonInit installs callbacks without routing a Window pointer through
// the original PE32 int ABI (nox_xxx_wndRadioButtonSetAllFn_4A87E0).
func RadioButtonInit(win *Window) {
	if win == nil {
		return
	}
	if win.Flags.Has(StatusImage) {
		win.SetAllFuncs(radioButtonProc, radioButtonDrawImg, nil)
	} else {
		win.SetAllFuncs(radioButtonProc, radioButtonDrawNoImg, nil)
	}
}

func radioButtonNotify(win *Window, code int, a1, a2 uintptr) {
	if parent := win.DrawData().Window; parent != nil {
		parent.Func94(AsWindowEvent(code, a1, a2))
	}
}

func radioButtonSelect(win *Window) {
	if parent := win.Parent(); parent != nil {
		group := win.DrawData().Group()
		for it := parent.Field100(); it != nil; it = it.Prev() {
			if it != win && it.DrawData().Group() == group {
				it.DrawData().Field0 &^= 0x4
			}
		}
	}
	win.DrawData().Field0 |= 0x4
}

func radioButtonProc(win *Window, e WindowEvent) WindowEventResp {
	switch e := e.(type) {
	case WindowKeyPress:
		switch e.Key {
		case keybind.KeyTab, keybind.KeyRight, keybind.KeyDown, keybind.KeyUp, keybind.KeyLeft:
			return RawEventResp(1)
		case keybind.KeyEnter, keybind.KeySpace:
			if e.Pressed && win.DrawData().Field0&0x4 == 0 {
				radioButtonNotify(win, 0x4007, uintptr(unsafe.Pointer(win)), 0)
				radioButtonSelect(win)
			}
			return RawEventResp(1)
		default:
			return RawEventResp(0)
		}
	case *WindowMouseState:
		switch e.State {
		case input.NOX_MOUSE_LEFT_DOWN:
			return RawEventResp(1)
		case input.NOX_MOUSE_LEFT_DRAG_END, input.NOX_MOUSE_LEFT_UP:
			if win.DrawData().Field0&0x4 != 0 {
				if win.DrawData().Field0&0x2 != 0 {
					return RawEventResp(1)
				}
				return RawEventResp(0)
			}
			a1, _ := e.EventArgsC()
			radioButtonNotify(win, 0x4007, uintptr(unsafe.Pointer(win)), a1)
			radioButtonSelect(win)
			return RawEventResp(1)
		case input.NOX_MOUSE_LEFT_PRESSED:
			a1, _ := e.EventArgsC()
			radioButtonNotify(win, 0x4000, uintptr(unsafe.Pointer(win)), a1)
			return RawEventResp(1)
		default:
			return RawEventResp(0)
		}
	default:
		switch e.EventCode() {
		case 17:
			if win.DrawData().Style.Has(StyleMouseTrack) {
				win.DrawData().Field0 |= 0x2
				a1, _ := e.EventArgsC()
				radioButtonNotify(win, 0x4005, uintptr(unsafe.Pointer(win)), a1)
				win.Focus()
			}
			return RawEventResp(1)
		case 18:
			if win.DrawData().Style.Has(StyleMouseTrack) {
				win.DrawData().Field0 &^= 0x2
				a1, _ := e.EventArgsC()
				radioButtonNotify(win, 0x4006, uintptr(unsafe.Pointer(win)), a1)
			}
			return RawEventResp(1)
		default:
			return RawEventResp(0)
		}
	}
}

func radioButtonCentered(win *Window) bool {
	data := (*RadioButtonData)(win.WidgetData)
	return data != nil && data.Field0 != 0
}

func radioButtonDrawText(win *Window, draw *WindowData, imageMode bool) {
	textCl := draw.TextColor()
	if textCl.Color32() == noxcolor.Transparent32RGBA5551 {
		return
	}
	r := win.GUI().Render()
	font := draw.Font()
	pos := win.GlobalPos()
	var pt image.Point
	if radioButtonCentered(win) {
		tw := r.GetStringSizeWrapped(font, draw.Text(), 0).X
		pt = image.Pt(pos.X+win.SizeVal.X/2-tw/2, pos.Y+win.SizeVal.Y/2-r.FontHeight(font)/2)
	} else if imageMode {
		pt = image.Pt(pos.X+28, pos.Y+(win.SizeVal.Y-r.FontHeight(font))/2)
	} else {
		pt = image.Pt(pos.X+18, pos.Y+win.SizeVal.Y/2-r.FontHeight(font)/2)
	}
	if win.Flags.Has(StatusSmoothText) {
		r.SetTextSmooting(true)
	}
	defer r.SetTextSmooting(false)
	r.Data().SetTextColor(textCl)
	r.DrawStringWrapped(font, draw.Text(), image.Rectangle{Min: pt, Max: pt.Add(image.Pt(win.SizeVal.X, 0))})
}

func radioButtonDrawNoImg(win *Window, draw *WindowData) int {
	r := win.GUI().Render()
	borderCl := draw.EnabledColor()
	backCl := draw.BackgroundColor()
	if win.Flags.IsEnabled() {
		if draw.Field0&0x2 != 0 {
			borderCl = draw.HighlightColor()
		}
	} else {
		backCl = draw.DisabledColor()
	}
	pos := win.GlobalPos()
	box := image.Pt(pos.X+4, pos.Y+win.SizeVal.Y/2-5)
	if backCl.Color32() != noxcolor.Transparent32RGBA5551 {
		r.DrawRectFilledOpaque(box.X, box.Y, 10, 10, backCl)
	}
	if borderCl.Color32() != noxcolor.Transparent32RGBA5551 {
		r.DrawBorder(box.X, box.Y, 10, 10, borderCl)
	}
	if draw.Field0&0x4 != 0 {
		if selectedCl := draw.SelectedColor(); selectedCl.Color32() != noxcolor.Transparent32RGBA5551 {
			r.DrawRectFilledOpaque(box.X+1, box.Y+1, 8, 8, selectedCl)
		}
	}
	radioButtonDrawText(win, draw, false)
	return 1
}

func radioButtonDrawImg(win *Window, draw *WindowData) int {
	r := win.GUI().Render()
	fg := r.Bag.AsImage(draw.EnImageHnd)
	bg := r.Bag.AsImage(draw.BgImageHnd)
	if win.Flags.IsEnabled() {
		if draw.Field0&0x2 != 0 {
			fg = r.Bag.AsImage(draw.HlImageHnd)
		}
	} else {
		bg = r.Bag.AsImage(draw.DisImageHnd)
	}
	pos := win.GlobalPos().Add(draw.ImgPtVal)
	if bg != nil {
		r.DrawImage16(bg, pos)
	}
	if draw.Field0&0x4 != 0 {
		if selected := r.Bag.AsImage(draw.SelImageHnd); selected != nil {
			r.DrawImage16(selected, pos)
		}
	} else if fg != nil {
		r.DrawImage16(fg, pos)
	}
	radioButtonDrawText(win, draw, true)
	return 1
}
