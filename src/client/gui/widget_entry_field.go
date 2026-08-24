package gui

import (
	"image"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"

	"github.com/opennox/libs/client/keybind"

	"github.com/opennox/opennox/v1/client/input"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type EntryFieldData struct {
	Text       [512]uint16
	Field_1024 uint32
	Field_1028 uint32
	Field_1032 uint32
	Field_1036 uint32
	Field_1040 uint16
	Field_1042 int16
	Field_1044 uint32
	Field_1048 uint32
	Field_1052 uint32
}

func (d *EntryFieldData) CWidgetData() unsafe.Pointer {
	return unsafe.Pointer(d)
}

// NewEntryFieldRaw is the native-width replacement for
// nox_gui_newEntryField_488500. The original PE32 constructor passes both the
// WindowData and Window pointers through C int values, which truncates them on
// every 64-bit target.
func NewEntryFieldRaw(g *GUI, parent *Window, status StatusFlags, px, py, w, h int, draw *WindowData, opts *EntryFieldData) *Window {
	if g == nil || draw == nil || opts == nil || !draw.Style.IsEntryField() {
		return nil
	}
	win := g.NewWindowRaw(parent, status, px, py, w, h, entryFieldProcPre)
	if win == nil {
		return nil
	}
	win.SetAllFuncs(EntryFieldProc, entryFieldDraw, nil)
	if draw.Window == nil {
		draw.Window = win
	}
	win.CopyDrawData(draw)

	d, _ := alloc.New(EntryFieldData{})
	*d = *opts
	clear(d.Text[:])
	if d.Field_1040 == 0 || d.Field_1040 > 256 {
		d.Field_1040 = 256
	}
	d.Field_1048 = 0
	d.Field_1052 = 0
	win.WidgetData = unsafe.Pointer(d)
	return win
}

func entryFieldData(win *Window) *EntryFieldData {
	if win == nil || win.WidgetData == nil {
		return nil
	}
	return (*EntryFieldData)(win.WidgetData)
}

func entryFieldLen(d *EntryFieldData) int {
	if d == nil {
		return 0
	}
	return int(uint16(d.Field_1052))
}

func entryFieldSetLen(d *EntryFieldData, n int) {
	d.Field_1052 = d.Field_1052&0xffff0000 | uint32(uint16(n))
}

func entryFieldSetText(d *EntryFieldData, text string) {
	if d == nil {
		return
	}
	clear(d.Text[:])
	units := utf16.Encode([]rune(text))
	limit := min(int(d.Field_1040)-1, 255)
	if limit < 0 {
		limit = 0
	}
	n := min(len(units), limit)
	copy(d.Text[:n], units[:n])
	entryFieldSetLen(d, n)
}

func entryFieldAppend(d *EntryFieldData, ch uint16) bool {
	if d.Field_1028 != 0 && !unicode.IsDigit(rune(ch)) {
		return false
	}
	if d.Field_1032 != 0 && !unicode.IsLetter(rune(ch)) && !unicode.IsDigit(rune(ch)) {
		return false
	}
	n := entryFieldLen(d)
	limit := min(int(d.Field_1040)-1, 255)
	if n >= limit {
		return false
	}
	d.Text[n] = ch
	n++
	d.Text[n] = 0
	entryFieldSetLen(d, n)
	return true
}

// EntryFieldOnChar accepts SDL/IME text input for the currently focused edit
// control without relying on the original 32-bit global Window pointer.
func EntryFieldOnChar(win *Window, ch uint16) bool {
	d := entryFieldData(win)
	if d == nil || !win.DrawData().Style.IsEntryField() {
		return false
	}
	switch ch {
	case '\n', '\r':
		d.Field_1044 = 0
	case 7, '\b', '\t', 11, 12:
	default:
		d.Field_1044 = 1
		entryFieldAppend(d, ch)
	}
	return true
}

func entryFieldNotify(win *Window, code int, a1, a2 uintptr) {
	if parent := win.DrawData().Window; parent != nil {
		parent.Func94(AsWindowEvent(code, a1, a2))
	}
}

func entryFieldProcPre(win *Window, ev WindowEvent) WindowEventResp {
	d := entryFieldData(win)
	switch ev := ev.(type) {
	case WindowDestroy:
		if win.WidgetData != nil {
			alloc.FreePtr(win.WidgetData)
			win.WidgetData = nil
		}
		return RawEventResp(0)
	case WindowFocus:
		win.DrawData().Field0Set(0x2, bool(ev))
		if win.GUI().inp != nil {
			win.GUI().inp.SetTextInput(bool(ev))
		}
		a1, _ := ev.EventArgsC()
		entryFieldNotify(win, 0x4003, a1, uintptr(win.ID()))
		return RawEventResp(1)
	}
	if d == nil {
		return RawEventResp(0)
	}
	a1, _ := ev.EventArgsC()
	switch ev.EventCode() {
	case 0x401D:
		return RawEventResp(uintptr(win.WidgetData))
	case 0x401E:
		if a1 == 0 {
			entryFieldSetText(d, "")
		} else {
			entryFieldSetText(d, alloc.GoString16((*uint16)(unsafe.Pointer(a1))))
		}
	}
	return RawEventResp(0)
}

// EntryFieldProc handles pointer-width-safe edit control input. It is exported
// so legacy dialog and console callbacks can delegate to it without re-entering
// the PE32 callback.
func EntryFieldProc(win *Window, ev WindowEvent) WindowEventResp {
	d := entryFieldData(win)
	if d == nil {
		return RawEventResp(0)
	}
	switch ev := ev.(type) {
	case *WindowMouseState:
		switch ev.State {
		case input.NOX_MOUSE_LEFT_UP:
			win.DrawData().Field0Set(0x2, true)
			win.Focus()
			return RawEventResp(1)
		case input.NOX_MOUSE_LEFT_PRESSED:
			if win.GetFlags().Has(StatusTabStop) {
				entryFieldNotify(win, 0x4000, uintptr(unsafe.Pointer(win)), 0)
			}
			return RawEventResp(1)
		}
	case *WindowMouseUnk:
		switch ev.Event {
		case 17:
			win.DrawData().Field0Set(0x2, true)
			entryFieldNotify(win, 0x4005, uintptr(unsafe.Pointer(win)), 0)
			return RawEventResp(1)
		case 18:
			win.DrawData().Field0Set(0x2, false)
			entryFieldNotify(win, 0x4006, uintptr(unsafe.Pointer(win)), 0)
			return RawEventResp(1)
		}
	case WindowKeyPress:
		if !ev.Pressed {
			return RawEventResp(1)
		}
		if ev.Key == keybind.KeyEnter {
			if d.Field_1044 == 0 {
				entryFieldNotify(win, 0x401F, uintptr(unsafe.Pointer(win)), 0)
			}
			return RawEventResp(1)
		}
		if ev.Key == keybind.KeyTab {
			return RawEventResp(1)
		}
		if win.GUI().inp == nil {
			return RawEventResp(1)
		}
		ch := win.GUI().inp.KeyToWChar(ev.Key)
		if ch == 0 {
			return RawEventResp(1)
		}
		n := entryFieldLen(d)
		if ch == '\b' {
			if n > 0 {
				n--
				d.Text[n] = 0
				entryFieldSetLen(d, n)
			}
			return RawEventResp(1)
		}
		entryFieldAppend(d, ch)
		return RawEventResp(1)
	}
	return RawEventResp(0)
}

func entryFieldDraw(win *Window, draw *WindowData) int {
	r := win.GUI().Render()
	if r == nil {
		return 1
	}
	d := entryFieldData(win)
	if d == nil {
		return 1
	}
	pos, sz := win.GlobalPos(), win.Size()
	if win.GetFlags().Has(StatusSmoothText) {
		r.SetTextSmooting(true)
		defer r.SetTextSmooting(false)
	}
	if win.GetFlags().Has(StatusImage) {
		img := draw.BackgroundImage()
		if !win.GetFlags().IsEnabled() {
			img = draw.DisabledImage()
		}
		if img != nil {
			r.DrawImage16(img, pos.Add(draw.ImagePoint()))
		}
	} else {
		bg := draw.BackgroundColor()
		if !win.GetFlags().IsEnabled() {
			bg = draw.DisabledColor()
		}
		if _, _, _, a := bg.RGBA(); a != 0 {
			r.DrawRectFilledOpaque(pos.X, pos.Y, sz.X, sz.Y, bg)
		}
		border := draw.EnabledColor()
		if draw.Field0&0x2 != 0 {
			border = draw.HighlightColor()
		}
		if _, _, _, a := border.RGBA(); a != 0 {
			r.DrawBorder(pos.X, pos.Y, sz.X, sz.Y, border)
		}
	}

	font := draw.Font()
	fh := r.FontHeight(font)
	y := pos.Y + (sz.Y-fh)/2
	x := pos.X + 5
	if label := draw.Text(); label != "" {
		r.Data().SetTextColor(draw.TextColor())
		r.DrawStringWrapped(font, label, image.Rect(pos.X+2, y, pos.X+sz.X, y+fh))
		x += r.GetStringSizeWrapped(font, label, 0).X + 6
	}
	text := alloc.GoString16(&d.Text[0])
	if d.Field_1024 != 0 {
		text = strings.Repeat("*", len([]rune(text)))
	}
	for text != "" && r.GetStringSizeWrapped(font, text, 0).X+10 > pos.X+sz.X-x {
		_, n := utf8.DecodeRuneInString(text)
		text = text[n:]
	}
	r.Data().SetTextColor(draw.TextColor())
	r.DrawStringWrapped(font, text, image.Rect(x, y, pos.X+sz.X-2, y+fh))
	if win.GUI().Focused() == win {
		cx := x + r.GetStringSizeWrapped(font, text, 0).X
		r.DrawRectFilledOpaque(cx, y, 2, fh, draw.TextColor())
	}
	return 1
}
