package gui

import (
	"fmt"
	"image"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const progressBarSetValueEvent = 0x4020

type progressBarData struct {
	value int32
}

func (d *progressBarData) CWidgetData() unsafe.Pointer {
	return unsafe.Pointer(d)
}

// NewProgressBarRaw is the native-width replacement for
// nox_gui_newProgressBar_4CAF10. The PE32 constructor passes the parent and
// the newly allocated Window through int values, truncating both on 64-bit
// hosts.
func NewProgressBarRaw(g *GUI, parent *Window, status StatusFlags, px, py, w, h int, draw *WindowData) *Window {
	if g == nil || draw == nil || !draw.Style.IsProgressBar() {
		return nil
	}
	d, _ := alloc.New(progressBarData{})
	win := g.NewWindowRaw(parent, status&^StatusEnabled, px, py, w, h, progressBarProc)
	if win == nil {
		return nil
	}
	win.WidgetData = unsafe.Pointer(d)
	if status.Has(StatusImage) {
		win.SetAllFuncs(nil, progressBarDrawImage, nil)
	} else {
		win.SetAllFuncs(nil, progressBarDraw, nil)
	}
	if draw.Window == nil {
		draw.Window = win
	}
	win.CopyDrawData(draw)
	return win
}

func progressBarState(win *Window) *progressBarData {
	if win == nil || win.WidgetData == nil {
		return nil
	}
	return (*progressBarData)(win.WidgetData)
}

func progressBarValue(win *Window) int {
	d := progressBarState(win)
	if d == nil {
		return 0
	}
	return int(d.value)
}

func progressBarProc(win *Window, ev WindowEvent) WindowEventResp {
	if ev.EventCode() != progressBarSetValueEvent {
		return nil
	}
	a1, _ := ev.EventArgsC()
	value := int(a1)
	if value >= 0 && value <= 100 {
		if d := progressBarState(win); d != nil {
			d.value = int32(value)
		}
	}
	return nil
}

func progressBarDraw(win *Window, draw *WindowData) int {
	r := win.GUI().Render()
	if r == nil {
		return 1
	}
	pos, sz := win.GlobalPos(), win.Size()
	if bg := draw.BackgroundColor(); bg.Color32() != noxcolor.Transparent32RGBA5551 {
		r.DrawRectFilledOpaque(pos.X, pos.Y, sz.X, sz.Y, bg)
	}
	if fill := draw.HighlightColor(); fill.Color32() != noxcolor.Transparent32RGBA5551 {
		width := sz.X * progressBarValue(win) / 100
		r.DrawRectFilledOpaque(pos.X, pos.Y, width, sz.Y, fill)
	}
	if textColor := draw.TextColor(); textColor.Color32() != noxcolor.Transparent32RGBA5551 {
		if win.GetFlags().Has(StatusSmoothText) {
			r.SetTextSmooting(true)
			defer r.SetTextSmooting(false)
		}
		text := fmt.Sprintf("%d%%", progressBarValue(win))
		font := draw.Font()
		textSize := r.GetStringSizeWrapped(font, text, 0)
		textPos := image.Pt(pos.X+(sz.X-textSize.X)/2, pos.Y+(sz.Y-textSize.Y)/2+1)
		r.Data().SetTextColor(textColor)
		r.DrawStringWrapped(font, text, image.Rectangle{Min: textPos, Max: textPos.Add(image.Pt(sz.X, 0))})
	}
	if border := draw.EnabledColor(); border.Color32() != noxcolor.Transparent32RGBA5551 {
		r.DrawBorder(pos.X, pos.Y, sz.X, sz.Y, border)
	}
	return 1
}

func progressBarDrawImage(win *Window, draw *WindowData) int {
	r := win.GUI().Render()
	if r == nil {
		return 1
	}
	width := win.Size().X * progressBarValue(win) / 100
	img := r.Bag.AsImage(draw.BgImageHnd)
	if img == nil || width <= 0 {
		return 1
	}

	data := r.Data()
	oldClip := data.ClipRect()
	oldClip2 := data.ClipRect2()
	oldUseClip := data.Clip()
	defer func() {
		data.SetClip(oldUseClip)
		data.SetClipRect(oldClip)
		data.SetClipRect2(oldClip2)
	}()

	pos := win.GlobalPos()
	clip := image.Rect(pos.X, pos.Y, pos.X+width, pos.Y+win.Size().Y).Intersect(data.Rect3())
	if clip.Empty() {
		return 1
	}
	data.SetClip(true)
	data.SetClipRect(clip)
	clip2 := clip
	clip2.Max = clip2.Max.Sub(image.Pt(1, 1))
	data.SetClipRect2(clip2)
	r.DrawImage16(img, pos)
	return 1
}
