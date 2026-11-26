package legacy

/*
#include <stdlib.h>
#include "GAME2_2.h"

extern void* dword_5d4594_1193348;
extern void* dword_5d4594_1193360;
*/
import "C"

import (
	"image"
	"unicode/utf16"
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const entryFieldPE32 = unsafe.Sizeof(uintptr(0)) == 4

var entryFieldIMEWindow *nox_window

//export nox_xxx_wndEditProc_487D70
func nox_xxx_wndEditProc_487D70(win *nox_window, event C.int, a3, a4 C.uintptr_t) C.int {
	return C.int(nox_xxx_wndEditProc_487D70_go(win, event, a3, a4))
}

//export nox_gui_newEntryField_488500
func nox_gui_newEntryField_488500(par *nox_window, flags, x, y, w, h C.int, draw *C.nox_window_data, opts unsafe.Pointer) *nox_window {
	if entryFieldPE32 {
		return newEntryFieldPE32(asWindow(par), gui.StatusFlags(flags), int(x), int(y), int(w), int(h),
			asWindowData(draw), (*gui.EntryFieldData)(opts))
	}
	return nox_gui_newEntryField_488500_go(par, flags, x, y, w, h, draw, opts)
}

func newEntryFieldPE32(parent *gui.Window, flags gui.StatusFlags, x, y, width, height int, draw *gui.WindowData, opts *gui.EntryFieldData) *nox_window {
	if draw == nil || opts == nil || !draw.Style.IsEntryField() {
		return nil
	}
	var g *gui.GUI
	if parent != nil {
		g = parent.GUI()
	} else {
		g = GetClient().Cli().GUI
	}
	win := g.NewWindowRaw(parent, flags, x, y, width, height, gui.WrapFuncC(C.nox_xxx_wndEditProcPre_488710))
	if win == nil {
		return nil
	}
	native := (*nox_window)(win.C())
	nox_xxx_wndEdit_488830(native)
	if draw.Window == nil {
		draw.Window = win
	}
	win.CopyDrawData(draw)
	clear(opts.Text[:128])
	clear(opts.Text[256:384])
	opts.Field_1044, opts.Field_1052 = 0, 0
	if int16(opts.Field_1040) >= 256 {
		opts.Field_1040 = 256
	}
	data, _ := alloc.New(*opts)
	data.Field_1048 = 0
	win.WidgetData = data.CWidgetData()

	if lang := nox_strman_get_lang_code(); lang == 6 || lang == 8 {
		listDraw := gui.WindowData{
			Style: 288, EnColorVal: draw.TextColorVal, HlColorVal: 0x80000000,
			SelColorVal: 0x80000000, TextColorVal: draw.TextColorVal,
		}
		listData := gui.ScrollListBoxData{Count: 128, Line_height: 10, Field_2: 1, Field_3: 1}
		list := Nox_gui_newScrollListBox_4A4310(nil, 17584, 0, height, 110, 119, &listDraw, &listData)
		if list == nil {
			win.Destroy()
			if draw.Window == win {
				draw.Window = nil
			}
			return nil
		}
		data.Field_1048 = uint32(uintptr(list.C()))
		list.Flags &^= gui.StatusImage
		gui.ScrollListBoxInit(list)
		nox_xxx_wndSetRectColor2MB_46AFE0((*nox_window)(list.C()), int32(nox_color_rgb_4344A0(0, 0, 0)))
	}
	return native
}

func entryFieldIMEListPE32(data *gui.EntryFieldData) *gui.Window {
	if data == nil || data.Field_1048 == 0 {
		return nil
	}
	// The PE32 widget data holds this pointer in a four-byte slot.
	return asWindow(*(**nox_window)(unsafe.Pointer(&data.Field_1048)))
}

//export nox_xxx_wndEditProcPre_488710
func nox_xxx_wndEditProcPre_488710(native *nox_window, event C.uint, a3, a4 C.uintptr_t) C.uintptr_t {
	if !entryFieldPE32 {
		return nox_xxx_wndEditProcPre_488710_go(native, event, a3, a4)
	}
	if native == nil {
		return 0
	}
	win := asWindow(native)
	data := (*gui.EntryFieldData)(win.WidgetData)
	switch event {
	case 2:
		if list := entryFieldIMEListPE32(data); list != nil {
			list.Destroy()
		}
		if data != nil {
			alloc.FreePtr(win.WidgetData)
			win.WidgetData = nil
		}
		if entryFieldIMEWindow == native {
			entryFieldIMEWindow = nil
		}
	case 23:
		if a3 != 0 {
			entryFieldIMEWindow = native
			nox_input_enableTextEdit_5700CA()
			win.DrawData().Field0 |= 6
		} else {
			nox_input_disableTextEdit_5700F6()
			entryFieldIMEWindow = nil
			win.DrawData().Field0 &^= 6
			if data != nil {
				if list := entryFieldIMEListPE32(data); list != nil {
					list.Hide()
				}
				data.Text[256] = 0
				data.Field_1052 &= 0xffff
			}
		}
		if owner := win.DrawData().Window; owner != nil {
			owner.Func94(gui.AsWindowEvent(0x4003, uintptr(a3), uintptr(win.ID())))
		}
		return 1
	case 0x401D:
		return C.uintptr_t(uintptr(win.WidgetData))
	case 0x401E:
		if data != nil {
			n := 0
			if a3 != 0 {
				src := unsafe.Pointer(uintptr(a3))
				for n < 255 {
					unit := *(*uint16)(unsafe.Add(src, 2*n))
					if unit == 0 {
						break
					}
					data.Text[n] = unit
					n++
				}
			}
			data.Text[n] = 0
			data.Text[256] = 0
			data.Field_1052 = uint32(n)
		}
	}
	return 0
}

func entryFieldTextUnits(text string) []uint16 {
	units := utf16.Encode([]rune(text))
	return units[:min(alloc.StrLenS(units), 255)]
}

//export nox_xxx_wndEdit_488830
func nox_xxx_wndEdit_488830(win *nox_window) C.int {
	if win == nil {
		return 0
	}
	draw := C.nox_xxx_wndEditDrawNoImage_488160
	if asWindow(win).GetFlags().Has(gui.StatusImage) {
		draw = C.nox_xxx_wndEditDrawWithImage_488870
	}
	return C.int(nox_window_set_all_funcs(win, C.nox_xxx_wndEditProc_487D70, draw, nil))
}

//export nox_xxx_wndEditDrawNoImage_488160
func nox_xxx_wndEditDrawNoImage_488160(win *nox_window, draw *C.nox_window_data) C.int {
	if !entryFieldPE32 {
		return nox_xxx_wndEditDraw_488160_go(win, draw)
	}
	return drawEntryFieldPE32(asWindow(win), draw, false)
}

//export nox_xxx_wndEditDrawWithImage_488870
func nox_xxx_wndEditDrawWithImage_488870(win *nox_window, draw *C.nox_window_data) C.int {
	if !entryFieldPE32 {
		return nox_xxx_wndEditDraw_488160_go(win, draw)
	}
	return drawEntryFieldPE32(asWindow(win), draw, true)
}

func drawEntryFieldPE32(win *gui.Window, native *C.nox_window_data, imageBacked bool) C.int {
	const transparent = 0x80000000
	draw := asWindowData(native)
	data := (*gui.EntryFieldData)(win.WidgetData)
	if data == nil {
		return 1
	}
	data.Field_1044 = 0
	flags := win.GetFlags()
	if flags.Has(gui.StatusSmoothText) {
		nox_draw_enableTextSmoothing_43F670(1)
	}
	defer nox_draw_enableTextSmoothing_43F670(0)
	pos, size := win.GlobalPos(), win.Size()
	x, yTop, width, height := int32(pos.X), int32(pos.Y), int32(size.X), int32(size.Y)
	font := draw.FontPtr
	fontHeight := nox_xxx_guiFontHeightMB_43F320(font)
	y := yTop + height/2 - fontHeight/2
	if imageBacked {
		img := draw.BackgroundImage()
		if !flags.IsEnabled() {
			img = draw.DisabledImage()
		}
		if img != nil {
			nox_client_drawImageAt_47D2C0((*nox_video_bag_image_t)(img.C()), x, yTop)
		}
		nox_xxx_drawSetTextColor_434390(int32(draw.TextColorVal))
	}
	if native.text[0] != 0 {
		var labelWidth C.int
		nox_xxx_drawGetStringSize_43F840(font, &native.text[0], &labelWidth, nil, 0)
		if !imageBacked {
			nox_xxx_drawSetTextColor_434390(int32(draw.TextColorVal))
		}
		nox_xxx_drawStringWrap_43FAF0(font, &native.text[0], x+2, y, width, 0)
		x += int32(labelWidth) + 6
		width -= int32(labelWidth) + 6
	}
	if !imageBacked {
		if maxWidth := int32(data.Field_1042); maxWidth > 0 && width > maxWidth {
			x = int32(pos.X) + int32(size.X) - maxWidth
			width = maxWidth
		}
		background, border := draw.BgColorVal, draw.EnColorVal
		if !flags.IsEnabled() {
			background = draw.DisColorVal
		} else if draw.Field0&2 != 0 {
			border = draw.HlColorVal
		}
		if background != transparent {
			nox_client_drawSetColor_434460(int32(background))
			nox_client_drawRectFilledOpaque_49CE30(x+1, yTop+1, width-2, height-2)
		}
		if border != transparent {
			nox_client_drawSetColor_434460(int32(border))
			nox_client_drawBorderLines_49CC70(x, yTop, width, height)
		}
	}
	if draw.TextColorVal == transparent {
		return 1
	}

	var text [256]uint16
	count := min(alloc.StrLenS(data.Text[:256]), 255)
	copy(text[:count], data.Text[:count])
	if data.Field_1024 != 0 {
		for i := range count {
			text[i] = '*'
		}
	}
	composition := (*wchar2_t)(unsafe.Pointer(&data.Text[256]))
	var textWidth, compositionWidth C.int
	label := (*wchar2_t)(unsafe.Pointer(&text[0]))
	nox_xxx_drawGetStringSize_43F840(font, label, &textWidth, nil, 0)
	nox_xxx_drawGetStringSize_43F840(font, composition, &compositionWidth, nil, 0)
	if flags.Has(1<<14) && textWidth+compositionWidth > 0 && (imageBacked || width >= 10) {
		for i := 1; i <= count && int32(textWidth+compositionWidth)+10 > width; i++ {
			label = (*wchar2_t)(unsafe.Pointer(&text[i]))
			nox_xxx_drawGetStringSize_43F840(font, label, &textWidth, nil, 0)
		}
	}
	if !imageBacked {
		if list := entryFieldIMEListPE32(data); list != nil {
			list.SetPos(image.Pt(int(x)+int(textWidth), int(y+fontHeight)))
		}
		width = 0
	}
	nox_xxx_drawSetTextColor_434390(int32(draw.TextColorVal))
	nox_xxx_drawStringWrap_43FAF0(font, label, x+5, y, width, 0)
	nox_xxx_drawSetTextColor_434390(int32(nox_color_rgb_4344A0(192, 0, 192)))
	nox_xxx_drawStringWrap_43FAF0(font, composition, x+int32(textWidth)+5, y, width, 0)
	if win.GUI().Focused() == win {
		blink := memmap.PtrUint8(0x5D4594, 1193344)
		on := *blink&8 != 0
		*blink++
		if on {
			nox_client_drawSetColor_434460(int32(draw.TextColorVal))
			nox_client_drawRectFilledOpaque_49CE30(x+int32(textWidth+compositionWidth)+5, y, 2, fontHeight)
		}
	}
	return 1
}

//export sub_488B60
func sub_488B60() C.int {
	if C.dword_5d4594_1193348 == nil {
		C.dword_5d4594_1193348 = C.calloc(1, 4)
	}
	return 1
}

//export sub_488BA0
func sub_488BA0() C.int {
	C.free(C.dword_5d4594_1193348)
	C.dword_5d4594_1193348 = nil
	return 1
}

//export nox_xxx_onChar_488BD0
func nox_xxx_onChar_488BD0(ch C.ushort) {
	native := entryFieldIMEWindow
	if C.dword_5d4594_1193348 == nil || native == nil {
		return
	}
	if lang := nox_strman_get_lang_code(); lang != 6 && lang != 8 {
		return
	}
	data := (*gui.EntryFieldData)(asWindow(native).WidgetData)
	if data == nil || data.Field_1036 != 0 || data.Field_1032 != 0 || data.Field_1028 != 0 {
		return
	}
	data.Field_1044 = 1
	switch ch {
	case 7, '\b', '\t', 11, 12:
	case '\n', '\r':
		data.Field_1044 = 0
	default:
		n := int(uint16(data.Field_1052))
		if n < min(int(int16(data.Field_1040))-1, 255) {
			data.Text[n], data.Text[n+1] = uint16(ch), 0
			n++
		}
		units := entryFieldTextUnits(GetClient().GetTextEditBuf())
		composition := data.Text[256:]
		copy(composition, units)
		composition[len(units)] = 0
		data.Field_1052 = uint32(len(units))<<16 | uint32(n)
		if list := entryFieldIMEListPE32(data); list != nil {
			list.Hide()
		}
	}
}

//export sub_4896E0
func sub_4896E0() C.int {
	C.free(C.dword_5d4594_1193360)
	C.dword_5d4594_1193360 = nil
	return 1
}
