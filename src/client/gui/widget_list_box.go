package gui

import (
	"image"
	"image/color"
	"strings"
	"unsafe"

	"github.com/opennox/libs/client/keybind"
	noxcolor "github.com/opennox/libs/color"

	"github.com/opennox/opennox/v1/client/input"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type ScrollListBoxItem struct {
	Field_0   uint32      // 0, 0; cumulative bottom edge in pixels
	Text      [256]uint16 // 1, 4
	Field_129 uint32      // 129, 516; text color
	Field_130 uint32      // 130, 520; rendered line height
}

type ScrollListBoxData struct {
	Count       uint16             // 0, 0
	Line_height uint16             // 0, 2
	Field_1     uint32             // 1, 4; keep the inserted line visible
	Field_2     uint32             // 2, 8; overwrite when full
	Field_3     uint32             // 3, 12; constructor-owned scrollbar
	Field_4     uint32             // 4, 16; multiple selection
	Field_5     uint32             // 5, 20; selection required
	Items       *ScrollListBoxItem // 6, 24
	Field_7     unsafe.Pointer     // 7, 28; up button
	Field_8     unsafe.Pointer     // 8, 32; down button
	Field_9     unsafe.Pointer     // 9, 36; slider
	Field_10    uint32             // 10, 40; total content height
	Field_11_0  uint16             // 11, 44; number of populated items
	Field_11_1  uint16             // 11, 46; insertion position
	Field_12    *uint32            // 12, 48; selected indices for multiselect
	Field_13_0  uint16             // 13, 52; viewport height
	Field_13_1  uint16             // 13, 54; scroll offset
}

func (d *ScrollListBoxData) CWidgetData() unsafe.Pointer {
	return unsafe.Pointer(d)
}

type scrollListBoxExt struct {
	selection int
}

type ScrollListBoxAssets struct {
	LoadImage func(name string) *noxrender.Image
	UpText    string
	DownText  string
}

var scrollListBoxExts = make(map[*Window]*scrollListBoxExt)

// NewScrollListBoxRaw is the native-width replacement for
// nox_gui_newScrollListBox_4A4310. The original constructor and callbacks use
// PE32 byte offsets and pass Window and WindowData pointers through C int.
func NewScrollListBoxRaw(g *GUI, parent *Window, status StatusFlags, px, py, w, h int, draw *WindowData, opts *ScrollListBoxData, assetsArg ...ScrollListBoxAssets) *Window {
	if g == nil || draw == nil || opts == nil || !draw.Style.IsScrollListBox() {
		return nil
	}
	if g.Render() != nil {
		if fh := g.Render().FontHeight(g.Render().Fonts.AsFont(draw.FontC())); int(opts.Line_height) < fh {
			opts.Line_height = uint16(fh)
		}
	}
	if opts.Line_height == 0 {
		opts.Line_height = 1
	}
	win := g.NewWindowRaw(parent, status, px, py, w, h, scrollListBoxProcPre)
	if win == nil {
		return nil
	}
	win.SetAllFuncs(scrollListBoxProc, scrollListBoxDraw, nil)
	if draw.Window == nil {
		draw.Window = win
	}
	win.CopyDrawData(draw)

	d, _ := alloc.New(ScrollListBoxData{})
	*d = *opts
	if d.Count != 0 {
		items, _ := alloc.Make([]ScrollListBoxItem(nil), int(d.Count))
		d.Items = &items[0]
	}
	d.Field_10 = 0
	d.Field_11_0 = 0
	d.Field_11_1 = 0
	d.Field_13_0 = uint16(max(h-scrollListBoxTitleHeight(win), 0))
	d.Field_13_1 = 0
	if d.Field_4 != 0 && d.Count != 0 {
		selected, _ := alloc.Make([]uint32(nil), int(d.Count))
		for i := range selected {
			selected[i] = ^uint32(0)
		}
		d.Field_12 = &selected[0]
	} else {
		d.Field_12 = nil
	}
	win.WidgetData = unsafe.Pointer(d)
	scrollListBoxExts[win] = &scrollListBoxExt{selection: -1}
	if d.Field_3 != 0 {
		var assets ScrollListBoxAssets
		if len(assetsArg) != 0 {
			assets = assetsArg[0]
		}
		if !scrollListBoxInitControls(g, win, status, w, h, assets) {
			return nil
		}
	}
	return win
}

func scrollListBoxControlDraw(win *Window, style StyleFlags, imageNames [4]string, text string, assets ScrollListBoxAssets) WindowData {
	draw := WindowData{Window: win, Style: style}
	if win.Flags.Has(StatusImage) {
		if assets.LoadImage != nil {
			draw.SetBackgroundImage(assets.LoadImage(imageNames[0]))
			draw.SetHighlightImage(assets.LoadImage(imageNames[1]))
			draw.SetDisabledImage(assets.LoadImage(imageNames[2]))
			draw.SetSelectedImage(assets.LoadImage(imageNames[3]))
		}
		return draw
	}
	draw.SetBackgroundColor(color.Black)
	draw.SetDisabledColor(color.Black)
	draw.SetEnabledColor(noxcolor.RGB5551Color(240, 180, 42))
	if style.IsPushButton() {
		draw.SetHighlightColor(color.White)
		draw.SetSelectedColor(noxcolor.RGB5551Color(255, 255, 128))
		draw.SetTextColor(noxcolor.RGB5551Color(240, 180, 42))
		draw.SetText(text)
	} else {
		draw.SetHighlightColor(color.Black)
		draw.SetSelectedColor(noxcolor.RGB5551Color(240, 180, 42))
	}
	return draw
}

func scrollListBoxInitControls(g *GUI, win *Window, status StatusFlags, w, h int, assets ScrollListBoxAssets) bool {
	d := scrollListBoxData(win)
	if d == nil {
		return false
	}
	titleH := scrollListBoxTitleHeight(win)
	innerH := h - titleH
	buttonH := 10
	sliderW := 10
	if status.Has(StatusImage) {
		buttonH = 13
		sliderW = 9
	}
	childStatus := (status &^ (StatusHidden | StatusBorder)) | StatusActive | StatusEnabled

	upDraw := scrollListBoxControlDraw(win, StylePushButton,
		[4]string{"DefaultLBUpButton", "DefaultLBUpButtonLit", "DefaultLBUpButtonDis", "DefaultLBUpButtonLit"},
		assets.UpText, assets)
	up := NewButtonRaw(g, win, childStatus, w-10, titleH, 10, buttonH, &upDraw)
	if up == nil {
		return false
	}
	d.Field_7 = unsafe.Pointer(up)

	downDraw := scrollListBoxControlDraw(win, StylePushButton,
		[4]string{"DefaultLBDownButton", "DefaultLBDownButtonLit", "DefaultLBDownButtonDis", "DefaultLBDownButtonLit"},
		assets.DownText, assets)
	down := NewButtonRaw(g, win, childStatus, w-10, titleH+innerH-buttonH, 10, buttonH, &downDraw)
	if down == nil {
		return false
	}
	d.Field_8 = unsafe.Pointer(down)

	sliderDraw := scrollListBoxControlDraw(win, StyleVertSlider,
		[4]string{"DefaultSliderThumb", "DefaultSliderThumbLit", "DefaultSliderThumbDis", "DefaultSliderThumbLit"},
		"", assets)
	slider := NewSliderRaw(g, win, childStatus, w-sliderW, titleH+buttonH, sliderW, innerH-2*buttonH,
		&sliderDraw, &SliderData{})
	if slider == nil {
		return false
	}
	d.Field_9 = unsafe.Pointer(slider)
	return true
}

func scrollListBoxData(win *Window) *ScrollListBoxData {
	if win == nil || win.WidgetData == nil {
		return nil
	}
	return (*ScrollListBoxData)(win.WidgetData)
}

func scrollListBoxItems(d *ScrollListBoxData) []ScrollListBoxItem {
	if d == nil || d.Items == nil || d.Count == 0 {
		return nil
	}
	return unsafe.Slice(d.Items, int(d.Count))
}

func scrollListBoxSelection(win *Window) int {
	if ext := scrollListBoxExts[win]; ext != nil {
		return ext.selection
	}
	return -1
}

func scrollListBoxSetSelection(win *Window, ind int) {
	ext := scrollListBoxExts[win]
	if ext == nil {
		ext = &scrollListBoxExt{selection: -1}
		scrollListBoxExts[win] = ext
	}
	ext.selection = ind
}

func scrollListBoxTitleHeight(win *Window) int {
	if win == nil || win.DrawData().Text() == "" || win.GUI().Render() == nil {
		return 0
	}
	return win.GUI().Render().FontHeight(win.DrawData().Font()) + 1
}

func scrollListBoxWindow(ptr unsafe.Pointer) *Window {
	return (*Window)(ptr)
}

func scrollListBoxNotify(win *Window, code int, a1, a2 uintptr) {
	if parent := win.DrawData().Window; parent != nil {
		parent.Func94(AsWindowEvent(code, a1, a2))
	}
}

func scrollListBoxSelected(d *ScrollListBoxData, win *Window, ind int) bool {
	if d.Field_4 == 0 {
		return scrollListBoxSelection(win) == ind
	}
	if d.Field_12 == nil {
		return false
	}
	for _, v := range unsafe.Slice(d.Field_12, int(d.Count)) {
		if int32(v) < 0 {
			break
		}
		if int(v) == ind {
			return true
		}
	}
	return false
}

func scrollListBoxClearSelection(d *ScrollListBoxData, win *Window) {
	if d.Field_4 == 0 {
		scrollListBoxSetSelection(win, -1)
		return
	}
	if d.Field_12 != nil {
		selected := unsafe.Slice(d.Field_12, int(d.Count))
		for i := range selected {
			selected[i] = ^uint32(0)
		}
	}
}

func scrollListBoxToggleSelection(d *ScrollListBoxData, win *Window, ind int) {
	if d.Field_4 == 0 {
		if ind == scrollListBoxSelection(win) && d.Field_5 == 0 {
			ind = -1
		}
		if ind < 0 && d.Field_5 != 0 {
			return
		}
		scrollListBoxSetSelection(win, ind)
		return
	}
	if d.Field_12 == nil || ind < 0 {
		return
	}
	sel := unsafe.Slice(d.Field_12, int(d.Count))
	for i, v := range sel {
		if int32(v) < 0 {
			sel[i] = uint32(ind)
			if i+1 < len(sel) {
				sel[i+1] = ^uint32(0)
			}
			return
		}
		if int(v) == ind {
			copy(sel[i:], sel[i+1:])
			sel[len(sel)-1] = ^uint32(0)
			return
		}
	}
}

func scrollListBoxIndexAt(win *Window, pos image.Point) int {
	d := scrollListBoxData(win)
	if d == nil {
		return -1
	}
	top := win.GlobalPos().Y + scrollListBoxTitleHeight(win)
	rel := pos.Y - top + int(d.Field_13_1)
	if rel < int(d.Field_13_1) || rel >= int(d.Field_13_1)+int(d.Field_13_0) {
		return -1
	}
	items := scrollListBoxItems(d)
	for i := 0; i < int(d.Field_11_0); i++ {
		if rel < int(items[i].Field_0) {
			return i
		}
	}
	return -1
}

func scrollListBoxSetScroll(win *Window, off int) {
	d := scrollListBoxData(win)
	if d == nil {
		return
	}
	maxOff := max(int(d.Field_10)-int(d.Field_13_0)+1, 0)
	off = min(max(off, 0), maxOff)
	d.Field_13_1 = uint16(min(off, int(^uint16(0))))
	if slider := scrollListBoxWindow(d.Field_9); slider != nil {
		sd := (*SliderData)(slider.WidgetData)
		if sd != nil {
			value := int(sd.Max) - off
			slider.Func94(AsWindowEvent(0x400A, uintptr(max(value, int(sd.Min))), 0))
		}
	}
}

func scrollListBoxScrollLines(win *Window, lines int) {
	d := scrollListBoxData(win)
	if d == nil || d.Field_11_0 == 0 {
		return
	}
	items := scrollListBoxItems(d)
	cur := 0
	for cur < int(d.Field_11_0) && int(items[cur].Field_0) <= int(d.Field_13_1) {
		cur++
	}
	cur = min(max(cur+lines, 0), int(d.Field_11_0)-1)
	off := 0
	if cur > 0 {
		off = int(items[cur-1].Field_0) + 1
	}
	scrollListBoxSetScroll(win, off)
}

func scrollListBoxReflow(win *Window) {
	d := scrollListBoxData(win)
	if d == nil {
		return
	}
	items := scrollListBoxItems(d)
	total := 0
	for i := 0; i < int(d.Field_11_0); i++ {
		total += int(items[i].Field_130) + 1
		items[i].Field_0 = uint32(total)
	}
	d.Field_10 = uint32(total)
	if slider := scrollListBoxWindow(d.Field_9); slider != nil {
		maxOff := max(total-int(d.Field_13_0)+3, 0)
		slider.Func94(AsWindowEvent(0x400B, 0, uintptr(maxOff)))
	}
	scrollListBoxSetScroll(win, int(d.Field_13_1))
}

func scrollListBoxAddLine(win *Window, str string, colorArg int) bool {
	d := scrollListBoxData(win)
	if d == nil || d.Count == 0 {
		return false
	}
	items := scrollListBoxItems(d)
	ind := min(int(d.Field_11_1), int(d.Field_11_0))
	if int(d.Field_11_0) >= int(d.Count) {
		if d.Field_2 == 0 {
			return false
		}
		ind = min(ind, len(items)-1)
		scrollListBoxNotify(win, 0x401B, 1, 0)
		copy(items[ind+1:], items[ind:len(items)-1])
	} else {
		copy(items[ind+1:int(d.Field_11_0)+1], items[ind:int(d.Field_11_0)])
		d.Field_11_0++
	}
	d.Field_11_1 = uint16(min(ind+1, int(d.Field_11_0)))
	item := &items[ind]
	*item = ScrollListBoxItem{}
	str = strings.TrimSuffix(str, "\n")
	alloc.StrCopyZero16(item.Text[:], str)
	item.Field_129 = win.DrawData().TextColorVal
	_ = colorArg
	height := int(d.Line_height)
	if r := win.GUI().Render(); r != nil {
		if win.Flags.Has(StatusOneLine) {
			height = r.FontHeight(win.DrawData().Font())
		} else if sz := r.GetStringSizeWrapped(win.DrawData().Font(), str, max(win.Size().X-7, 1)); sz.Y > height {
			height = sz.Y
		}
	}
	item.Field_130 = uint32(max(height, 1))
	scrollListBoxReflow(win)
	if d.Field_1 != 0 {
		for int(item.Field_0) >= int(d.Field_13_1)+int(d.Field_13_0) {
			old := d.Field_13_1
			scrollListBoxScrollLines(win, 1)
			if old == d.Field_13_1 {
				break
			}
		}
	}
	return true
}

func scrollListBoxDelete(win *Window, ind int) bool {
	d := scrollListBoxData(win)
	if d == nil || ind < 0 || ind >= int(d.Field_11_0) {
		return false
	}
	items := scrollListBoxItems(d)
	copy(items[ind:], items[ind+1:int(d.Field_11_0)])
	d.Field_11_0--
	d.Field_11_1 = d.Field_11_0
	items[d.Field_11_0] = ScrollListBoxItem{}
	if d.Field_4 == 0 {
		sel := scrollListBoxSelection(win)
		if sel == ind {
			scrollListBoxSetSelection(win, -1)
		} else if sel > ind {
			scrollListBoxSetSelection(win, sel-1)
		}
	} else {
		scrollListBoxClearSelection(d, win)
	}
	scrollListBoxReflow(win)
	return true
}

func scrollListBoxProcPre(win *Window, e WindowEvent) WindowEventResp {
	d := scrollListBoxData(win)
	if d == nil {
		return RawEventResp(0)
	}
	a1, a2 := e.EventArgsC()
	switch e := e.(type) {
	case WindowDestroy:
		delete(scrollListBoxExts, win)
		if d.Items != nil {
			alloc.FreePtr(unsafe.Pointer(d.Items))
			d.Items = nil
		}
		if d.Field_12 != nil {
			alloc.FreePtr(unsafe.Pointer(d.Field_12))
			d.Field_12 = nil
		}
		alloc.FreePtr(win.WidgetData)
		win.WidgetData = nil
		return RawEventResp(0)
	case WindowFocus:
		win.DrawData().Field0Set(0x2, bool(e))
		scrollListBoxNotify(win, 0x4003, a1, uintptr(win.ID()))
		return RawEventResp(1)
	case *StaticTextSetText:
		win.DrawData().SetText(e.Str)
		d.Field_13_0 = uint16(max(win.Size().Y-scrollListBoxTitleHeight(win), 0))
		return RawEventResp(0)
	}
	switch e.EventCode() {
	case 0x4000, 0x4007:
		ptr := unsafe.Pointer(a1)
		if ptr == d.Field_7 {
			scrollListBoxScrollLines(win, -1)
		} else if ptr == d.Field_8 {
			scrollListBoxScrollLines(win, 1)
		}
	case 0x4004:
		d.Field_13_0 = uint16(max(int(a2)-scrollListBoxTitleHeight(win), 0))
		scrollListBoxReflow(win)
	case 0x4009:
		if slider := scrollListBoxWindow(d.Field_9); slider != nil {
			if sd := (*SliderData)(slider.WidgetData); sd != nil {
				scrollListBoxSetScroll(win, int(sd.Max)-int(a2))
			}
		}
	case 0x400D:
		if scrollListBoxAddLine(win, alloc.GoString16((*uint16)(unsafe.Pointer(a1))), int(a2)) {
			return RawEventResp(1)
		}
	case 0x400E:
		scrollListBoxDelete(win, int(a1))
	case 0x400F:
		items := scrollListBoxItems(d)
		for i := range items {
			items[i] = ScrollListBoxItem{}
		}
		if a1 != 1 {
			d.Field_13_1 = 0
		}
		d.Field_10, d.Field_11_0, d.Field_11_1 = 0, 0, 0
		scrollListBoxClearSelection(d, win)
		scrollListBoxReflow(win)
	case 0x4012:
		if int(a1) < 0 {
			d.Field_11_1 = 0
		} else {
			d.Field_11_1 = uint16(min(int(a1), int(d.Field_11_0)))
		}
	case 0x4013:
		ind := int(a1)
		if ind < 0 || ind >= int(d.Field_11_0) || scrollListBoxItems(d)[ind].Text[0] == 0 {
			scrollListBoxClearSelection(d, win)
			return RawEventResp(0)
		}
		if d.Field_4 != 0 {
			scrollListBoxClearSelection(d, win)
			scrollListBoxToggleSelection(d, win, ind)
		} else {
			scrollListBoxSetSelection(win, ind)
		}
		item := scrollListBoxItems(d)[ind]
		if int(item.Field_0) < int(d.Field_13_1) {
			scrollListBoxSetScroll(win, int(item.Field_0)-int(item.Field_130)-1)
		} else if int(item.Field_0) > int(d.Field_13_1)+int(d.Field_13_0) {
			scrollListBoxSetScroll(win, int(item.Field_0)-int(d.Field_13_0))
		}
	case 0x4014:
		if d.Field_4 != 0 {
			return RawEventResp(uintptr(unsafe.Pointer(d.Field_12)))
		}
		return RawEventResp(uintptr(scrollListBoxSelection(win)))
	case 0x4015:
		if d.Field_4 != 0 {
			scrollListBoxToggleSelection(d, win, int(a1))
		}
	case 0x4016:
		ind := int(a1)
		if ind >= 0 && ind < int(d.Field_11_0) {
			return RawEventResp(uintptr(unsafe.Pointer(&scrollListBoxItems(d)[ind].Text[0])))
		}
	case 0x4017:
		ind := int(a2)
		if ind >= 0 && ind < int(d.Field_11_0) {
			alloc.StrCopyZero16(scrollListBoxItems(d)[ind].Text[:], alloc.GoString16((*uint16)(unsafe.Pointer(a1))))
		}
	case 0x4018:
		d.Field_7 = unsafe.Pointer(a1)
	case 0x4019:
		d.Field_8 = unsafe.Pointer(a1)
	case 0x401A:
		d.Field_9 = unsafe.Pointer(a1)
		scrollListBoxReflow(win)
	case 0x401B:
		ind := min(int(a1), int(d.Field_11_0))
		for ind > 0 {
			scrollListBoxDelete(win, 0)
			ind--
		}
	case 0x401C:
		ind := int(a1)
		off := 0
		if ind > 0 && ind <= int(d.Field_11_0) {
			off = int(scrollListBoxItems(d)[ind-1].Field_0) + 1
		}
		scrollListBoxSetScroll(win, off)
	}
	return RawEventResp(0)
}

func scrollListBoxMoveSelection(win *Window, delta int) {
	d := scrollListBoxData(win)
	if d == nil || d.Field_4 != 0 || d.Field_11_0 == 0 {
		return
	}
	ind := scrollListBoxSelection(win)
	if ind < 0 {
		ind = 0
	} else {
		ind = min(max(ind+delta, 0), int(d.Field_11_0)-1)
	}
	win.Func94(AsWindowEvent(0x4013, uintptr(ind), 0))
	scrollListBoxNotify(win, 0x4010, uintptr(unsafe.Pointer(win)), uintptr(ind))
}

func scrollListBoxProc(win *Window, e WindowEvent) WindowEventResp {
	d := scrollListBoxData(win)
	if d == nil {
		return RawEventResp(0)
	}
	switch e := e.(type) {
	case *WindowMouseState:
		switch e.State {
		case input.NOX_MOUSE_LEFT_DOWN:
			return RawEventResp(1)
		case input.NOX_MOUSE_LEFT_DRAG_END, input.NOX_MOUSE_LEFT_UP:
			ind := scrollListBoxIndexAt(win, e.Pos)
			scrollListBoxToggleSelection(d, win, ind)
			scrollListBoxNotify(win, 0x4010, uintptr(unsafe.Pointer(win)), uintptr(ind))
			return RawEventResp(1)
		case input.NOX_MOUSE_LEFT_PRESSED:
			if win.DrawData().Style.Has(StyleTabStop) {
				scrollListBoxNotify(win, 0x4000, uintptr(unsafe.Pointer(win)), 0)
			}
			return RawEventResp(1)
		}
	case *WindowMouseUnk:
		if e.Event == 17 || e.Event == 18 {
			return RawEventResp(1)
		}
	case WindowKeyPress:
		if !e.Pressed {
			return RawEventResp(1)
		}
		switch e.Key {
		case keybind.KeyUp:
			scrollListBoxMoveSelection(win, -1)
			return RawEventResp(1)
		case keybind.KeyDown:
			scrollListBoxMoveSelection(win, 1)
			return RawEventResp(1)
		case keybind.KeyTab, keybind.KeyEnter, keybind.KeySpace:
			return RawEventResp(1)
		}
	}
	return RawEventResp(0)
}

func scrollListBoxDraw(win *Window, draw *WindowData) int {
	r := win.GUI().Render()
	if r == nil {
		return 1
	}
	d := scrollListBoxData(win)
	if d == nil {
		return 1
	}
	pos := win.GlobalPos()
	w, h := win.Size().X, win.Size().Y
	if win.Flags.Has(StatusSmoothText) {
		r.SetTextSmooting(true)
		defer r.SetTextSmooting(false)
	}
	if win.Flags.Has(StatusImage) {
		img := draw.BackgroundImage()
		if !win.Flags.IsEnabled() {
			img = draw.DisabledImage()
		}
		if img != nil {
			r.DrawImage16(img, pos.Add(draw.ImagePoint()))
		}
	} else {
		bg := draw.BackgroundColor()
		if !win.Flags.IsEnabled() {
			bg = draw.DisabledColor()
		}
		if _, _, _, a := bg.RGBA(); a != 0 {
			r.DrawRectFilledOpaque(pos.X, pos.Y, w, h, bg)
		}
		border := draw.EnabledColor()
		if draw.Field0&0x2 != 0 {
			border = draw.HighlightColor()
		}
		if _, _, _, a := border.RGBA(); a != 0 {
			r.DrawBorder(pos.X, pos.Y, w, h, border)
		}
	}
	top := pos.Y
	font := draw.Font()
	if title := draw.Text(); title != "" {
		r.Data().SetTextColor(draw.TextColor())
		r.DrawStringWrapped(font, title, image.Rect(pos.X+1, top, pos.X+w, top+r.FontHeight(font)))
		top += r.FontHeight(font) + 1
	}
	items := scrollListBoxItems(d)
	for i := 0; i < int(d.Field_11_0); i++ {
		item := &items[i]
		end := int(item.Field_0)
		lineH := int(item.Field_130) + 1
		start := end - lineH
		if end <= int(d.Field_13_1) || start > int(d.Field_13_1)+int(d.Field_13_0) {
			continue
		}
		y := top + start - int(d.Field_13_1)
		if scrollListBoxSelected(d, win, i) {
			cl := draw.SelectedColor()
			if _, _, _, a := cl.RGBA(); a != 0 {
				r.DrawRectFilledOpaque(pos.X, y, w, lineH, cl)
			}
		}
		r.Data().SetTextColor(draw.TextColor())
		text := alloc.GoString16(&item.Text[0])
		r.DrawStringWrapped(font, text, image.Rect(pos.X+5, y+2, pos.X+w-2, y+lineH))
	}
	return 1
}
