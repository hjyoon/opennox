package legacy

/*
#include "game3_widgets_native.h"
*/
import "C"

import (
	"github.com/opennox/opennox/v1/client/gui"
	"image"
	"math"
	"unsafe"
)

// This is a scalar 16-byte record. Window, owner and thumb pointers are typed
// native fields on gui.Window, never appended to this record as DWORDs.
type game3SliderData struct {
	Min, Max int32
	Scale    float32
	Value    int32
}

func game3Slider(w *C.nox_window) *game3SliderData {
	if w == nil {
		return nil
	}
	return (*game3SliderData)(w.widget_data)
}
func game3WidgetEvent(w *C.nox_window, e int, a, b uintptr) uintptr {
	if w == nil {
		return 0
	}
	r := asWindow(w).Func94(gui.AsWindowEvent(e, a, b))
	if r == nil {
		return 0
	}
	return r.EventRespC()
}
func game3SliderThumb(w *C.nox_window, x, y int32) {
	if w != nil && w.field_100 != nil {
		asWindow(w.field_100).SetPos(image.Pt(int(x), int(y)))
	}
}
func game3WidgetNotify(w *C.nox_window, owner *C.nox_window, e int, value int32) {
	game3WidgetEvent(owner, e, uintptr(unsafe.Pointer(w)), uintptr(value))
}

// Finite, representable conversion matches C truncation toward zero. The
// legacy out-of-range/NaN conversion was undefined; native policy yields the
// x86 indefinite int rather than depending on the host's int width.
func game3Trunc(v float64) int32 {
	if math.IsNaN(v) || v >= 2147483648 || v < -2147483648 {
		return math.MinInt32
	}
	return int32(v)
}
func game3SliderKey(w *C.nox_window, d *game3SliderData, vertical bool, key uintptr, pressed uintptr) C.int {
	next, prev, inc, dec := uintptr(15), uintptr(200), uintptr(203), uintptr(205)
	if vertical {
		prev, inc, dec = 203, 200, 208
	}
	if key == next || (!vertical && key == 208) || (vertical && key == 205) {
		if pressed == 2 {
			C.nox_xxx_wndRetNULL_46A8A0()
		}
		return 1
	}
	if key == prev {
		if pressed == 2 {
			C.nox_xxx_wndRetNULL_0_46A8B0()
		}
		return 1
	}
	if key != inc && key != dec {
		return 0
	}
	if pressed != 2 || d == nil {
		return 1
	}
	if key == inc {
		if d.Value >= d.Max-1 {
			return 1
		}
		d.Value += 2
	} else {
		if d.Value <= d.Min+1 {
			return 1
		}
		d.Value -= 2
	}
	game3WidgetNotify(w, w.draw_data.win, 16393, d.Value)
	if vertical {
		game3SliderThumb(w, 0, game3Trunc(float64(d.Max-d.Value)*float64(d.Scale)))
	} else {
		game3SliderThumb(w, game3Trunc(float64(d.Value-d.Min)*float64(d.Scale)), 0)
	}
	return 1
}
func game3SliderInput(w *C.nox_window, vertical bool, e int, a, b uintptr) C.int {
	if w == nil {
		return 0
	}
	d := game3Slider(w)
	switch e {
	case 5:
		return 1
	case 6, 7:
		pos := asWindow(w).GlobalPos()
		p := int32(uint16(a))
		origin, size := int32(pos.X), int32(w.width)
		if vertical {
			p = int32(uint16(a >> 16))
			origin, size = int32(pos.Y), int32(w.height)
		}
		if p > size+origin-10 {
			p = size + origin - 10
		}
		if vertical {
			game3SliderThumb(w, 0, p-origin-5)
		} else {
			game3SliderThumb(w, p-origin-5, 0)
		}
		game3WidgetEvent(w, 0x4000, 0, a)
		return 1
	case 8:
		if uint32(w.draw_data.style)&0x100 != 0 {
			game3WidgetNotify(w, w.draw_data.win, 0x4000, 0)
		}
		return 1
	case 17, 18:
		if uint32(w.draw_data.style)&0x100 != 0 {
			color := w.draw_data.en_color
			event := 16390
			if e == 17 {
				color = w.draw_data.hl_color
				event = 16389
			}
			C.nox_xxx_wndSetRectColor2MB_46AFE0(w, C.int(color))
			game3WidgetNotify(w, w.draw_data.win, event, 0)
			if e == 17 {
				C.nox_xxx_windowFocus_46B500(w)
			}
		}
		return 1
	case 21:
		return game3SliderKey(w, d, vertical, a, b)
	}
	return 0
}

//export sub_4B4860
func sub_4B4860(w *C.nox_window, e C.int, a, b C.uintptr_t) C.int {
	return game3SliderInput(w, false, int(e), uintptr(a), uintptr(b))
}

//export nox_xxx_wndScrollBoxDraw_4B4BA0
func nox_xxx_wndScrollBoxDraw_4B4BA0(w *C.nox_window, e C.int, a, b C.uintptr_t) C.int {
	return game3SliderInput(w, true, int(e), uintptr(a), uintptr(b))
}

func game3SliderPre(w *C.nox_window, vertical bool, e int, a, b uintptr) C.int {
	if w == nil {
		return 0
	}
	d := game3Slider(w)
	if e == 2 {
		C.free(w.widget_data)
		w.widget_data = nil
		return 0
	}
	if d == nil {
		return 0
	}
	switch e {
	case 16394:
		value := int32(a)
		if value >= d.Min && value <= d.Max {
			d.Value = value
			if vertical {
				game3SliderThumb(w, 0, game3Trunc(float64(d.Max-value)*float64(d.Scale)))
			} else {
				game3SliderThumb(w, game3Trunc(float64(value)*float64(d.Scale)), 0)
			}
		}
		return 0
	case 16395:
		d.Min = int32(a)
		d.Max = int32(b)
		d.Value = d.Min
		delta := d.Max - d.Min
		d.Scale = 0
		size := int32(w.width)
		if vertical {
			size = int32(w.height)
		}
		if delta != 0 {
			d.Scale = float32(size-10) / float32(delta)
		}
		if vertical {
			game3SliderThumb(w, 0, game3Trunc(float64(delta)*float64(d.Scale)))
		} else {
			game3SliderThumb(w, 0, 0)
		}
		return 0
	case 0x4007:
		if vertical {
			game3WidgetNotify(w, w.parent, 16396, d.Value)
		}
		return 0
	case 23:
		if a != 0 {
			w.draw_data.field_0 |= 2
		} else {
			w.draw_data.field_0 &^= 2
		}
		game3WidgetEvent(w.parent, 16387, a, uintptr(asWindow(w).ID()))
		return 1
	case 0x4000:
		pos := asWindow(w).GlobalPos()
		cursor, origin, size := int32(uint16(b)), int32(pos.X), int32(w.width)
		if vertical {
			cursor, origin, size = int32(b>>16), int32(pos.Y), int32(w.height)
		}
		if cursor < origin {
			game3SliderThumb(w, 0, 0)
			if vertical {
				d.Value = d.Max
			} else {
				d.Value = d.Min
			}
		} else if cursor < origin+size {
			var value int32
			if d.Scale != 0 {
				value = game3Trunc(float64(cursor-origin) / float64(d.Scale))
			} else if !vertical {
				value = d.Min
			}
			if value > d.Max {
				value = d.Max
			}
			if vertical {
				d.Value = d.Max - value
			} else {
				d.Value = value
			}
		} else {
			if vertical {
				var thumb int32
				if w.field_100 != nil {
					thumb = int32(w.field_100.height)
				}
				game3SliderThumb(w, 0, size-thumb)
				d.Value = d.Min
			} else {
				var thumb int32
				if w.field_100 != nil {
					thumb = int32(w.field_100.width)
				}
				game3SliderThumb(w, size-thumb, 0)
				d.Value = d.Max
			}
		}
		game3WidgetNotify(w, w.parent, 16393, d.Value)
		return 0
	}
	return 0
}

//export sub_4B5010
func sub_4B5010(w *C.nox_window, e C.int, a, b C.uintptr_t) C.int {
	return game3SliderPre(w, false, int(e), uintptr(a), uintptr(b))
}

//export nox_xxx_wndScrollBoxProc_4B5320
func nox_xxx_wndScrollBoxProc_4B5320(w *C.nox_window, e C.int, a, b C.uintptr_t) C.int {
	return game3SliderPre(w, true, int(e), uintptr(a), uintptr(b))
}

func game3SliderInstall(w *C.nox_window, vertical bool) C.int {
	if w == nil {
		return 0
	}
	image := uint32(w.flags)&0x80 != 0
	var input gui.WindowFunc
	var draw gui.WindowDrawFunc
	if vertical {
		input = func(win *gui.Window, e gui.WindowEvent) gui.WindowEventResp {
			ev := e.EventCode()
			a, b := e.EventArgsC()
			return gui.RawEventResp(game3SliderInput((*C.nox_window)(win.C()), true, ev, a, b))
		}
		draw = func(win *gui.Window, d *gui.WindowData) int {
			if image {
				return 1
			}
			return int(game3SliderDraw((*C.nox_window)(win.C()), (*C.nox_window_data)(d.C()), true, false))
		}
	} else {
		input = func(win *gui.Window, e gui.WindowEvent) gui.WindowEventResp {
			ev := e.EventCode()
			a, b := e.EventArgsC()
			return gui.RawEventResp(game3SliderInput((*C.nox_window)(win.C()), false, ev, a, b))
		}
		draw = func(win *gui.Window, d *gui.WindowData) int {
			return int(game3SliderDraw((*C.nox_window)(win.C()), (*C.nox_window_data)(d.C()), false, image))
		}
	}
	asWindow(w).SetAllFuncs(input, draw, nil)
	return 0
}

//export sub_4B51A0
func sub_4B51A0(w *C.nox_window) C.int { return game3SliderInstall(w, false) }

//export nox_xxx_wndScrollBoxSetAllFn_4B5500
func nox_xxx_wndScrollBoxSetAllFn_4B5500(w *C.nox_window) C.int { return game3SliderInstall(w, true) }

func game3SliderDraw(w *C.nox_window, d *C.nox_window_data, vertical, image bool) C.int {
	if w == nil || d == nil {
		return 1
	}
	p := asWindow(w).GlobalPos()
	x, y, width, height := int32(p.X), int32(p.Y), int32(w.width), int32(w.height)
	if image {
		img := d.bg_image
		if uint32(w.flags)&8 == 0 {
			img = d.dis_image
		}
		if img != nil {
			C.nox_client_drawImageAt_47D2C0((*C.nox_video_bag_image_t)(img), C.int(x), C.int(y))
		}
		return 1
	}
	line, bg := uint32(d.en_color), uint32(d.bg_color)
	if uint32(w.flags)&8 != 0 {
		if uint32(d.field_0)&2 != 0 {
			if vertical {
				line = uint32(d.sel_color)
			}
			if uint32(d.hl_color) != 0x80000000 {
				C.nox_client_drawSetColor_434460(C.int(d.hl_color))
				C.nox_client_drawBorderLines_49CC70(C.int(x), C.int(y), C.int(width), C.int(height))
			}
		}
	} else {
		bg = uint32(d.dis_color)
	}
	if bg != 0x80000000 {
		C.nox_client_drawSetColor_434460(C.int(bg))
		C.nox_client_drawRectFilledOpaque_49CE30(C.int(x+1), C.int(y+1), C.int(width-2), C.int(height-2))
	}
	if line != 0x80000000 {
		C.nox_client_drawSetColor_434460(C.int(line))
		if vertical {
			C.nox_client_drawRectFilledOpaque_49CE30(C.int(x+width/2-1), C.int(y+4), 3, C.int(height-8))
		} else {
			C.nox_client_drawRectFilledOpaque_49CE30(C.int(x), C.int(y+height/2-1), C.int(width), 3)
		}
	}
	return 1
}

//export sub_4B51E0
func sub_4B51E0(w *C.nox_window, d *C.nox_window_data) C.int {
	return game3SliderDraw(w, d, false, false)
}

//export sub_4B52C0
func sub_4B52C0(w *C.nox_window, d *C.nox_window_data) C.int {
	return game3SliderDraw(w, d, false, true)
}

//export nox_xxx_wndScrollBoxDraw_4B5540
func nox_xxx_wndScrollBoxDraw_4B5540(w *C.nox_window, d *C.nox_window_data) C.int {
	return game3SliderDraw(w, d, true, false)
}

//export nox_xxx_wndScrollBoxDraw_4B5620
func nox_xxx_wndScrollBoxDraw_4B5620(w *C.nox_window, d *C.nox_window_data) C.int { return 1 }

//export nox_xxx_wndScrollBoxButtonCreate_4B5640
func nox_xxx_wndScrollBoxButtonCreate_4B5640(parent *C.nox_window, flags C.int, src *C.nox_window_data) *C.nox_window {
	if parent == nil || src == nil {
		return nil
	}
	// Temporary pointer-bearing descriptors use C storage; the constructor copies
	// their contents synchronously and keeps only the C-owned window pointers.
	p := (*C.nox_window_data)(C.calloc(1, C.sizeof_nox_window_data))
	if p == nil {
		return nil
	}
	defer C.free(unsafe.Pointer(p))
	p.text_color = src.text_color
	p.win = parent
	p.style = 1
	flags = C.int(uint32(flags)&^0x10 | 0x0c)
	if uint32(flags)&0x80 == 0 {
		p.sel_color = src.sel_color
		p.en_color = src.bg_color
		p.bg_color = src.en_color
	} else {
		p.en_image = src.en_image
		p.dis_image = src.dis_image
		p.sel_image = src.sel_image
		p.bg_image = src.bg_image
		p.hl_image = src.hl_image
	}
	width, height := C.int(parent.width), C.int(10)
	if uint32(src.style)&0x10 != 0 {
		width, height = 10, C.int(parent.height)
	}
	button := gui.NewButtonOrCheckboxRaw(asWindow(parent).GUI(), asWindow(parent), gui.StatusFlags(flags), 0, 0, int(width), int(height), asWindowData(p))
	if button == nil {
		return nil
	}
	return (*C.nox_window)(button.C())
}

//export nox_gui_newSlider_4B4EE0
func nox_gui_newSlider_4B4EE0(parent *C.nox_window, flags, x, y, width, height C.int, draw *C.nox_window_data, opts *C.float) *C.nox_window {
	if draw == nil || opts == nil {
		return nil
	}
	vertical := uint32(draw.style)&0x10 == 0
	if vertical && uint32(draw.style)&8 == 0 {
		return nil
	}
	var g *gui.GUI
	if parent != nil {
		g = asWindow(parent).GUI()
	} else if GetClient() != nil && GetClient().Cli() != nil {
		g = GetClient().Cli().GUI
	}
	if g == nil {
		return nil
	}
	pre := func(win *gui.Window, e gui.WindowEvent) gui.WindowEventResp {
		ev := e.EventCode()
		a, b := e.EventArgsC()
		return gui.RawEventResp(game3SliderPre((*C.nox_window)(win.C()), vertical, ev, a, b))
	}
	win := g.NewWindowRaw(asWindow(parent), gui.StatusFlags(uint32(flags)|0x100), int(x), int(y), int(width), int(height), pre)
	if win == nil {
		return nil
	}
	cw := (*C.nox_window)(win.C())
	game3SliderInstall(cw, vertical)
	if draw.win == nil {
		draw.win = cw
	}
	win.CopyDrawData(asWindowData(draw))
	nox_xxx_wndScrollBoxButtonCreate_4B5640(cw, C.int(uint32(flags)|0x100), draw)
	values := (*game3SliderData)(unsafe.Pointer(opts))
	if values.Max == values.Min {
		values.Max = values.Min + 1
	}
	extent := int32(width) - 10
	if vertical {
		extent = int32(height) - 10
	}
	values.Scale = float32(float64(extent) / float64(values.Max-values.Min))
	data := (*game3SliderData)(C.calloc(1, 16))
	if data == nil {
		win.Destroy()
		return nil
	}
	*data = *values
	cw.widget_data = unsafe.Pointer(data)
	// GUI disposal runs registered cleanup before the window returns to its
	// allocation pool. Explicit event 2 already frees and clears this slot.
	win.AddCleanup(func() {
		if cw.widget_data == unsafe.Pointer(data) {
			C.free(unsafe.Pointer(data))
			cw.widget_data = nil
		}
	})
	return cw
}
