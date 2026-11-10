package legacy

/*
#include "game3_widgets_native.h"
*/
import "C"

import (
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/memmap"
	"unsafe"
)

func game3ListData(w *C.nox_window) *gui.ScrollListBoxData {
	if w == nil {
		return nil
	}
	return (*gui.ScrollListBoxData)(w.widget_data)
}
func game3ListItems(d *gui.ScrollListBoxData) []gui.ScrollListBoxItem {
	if d == nil || d.Items == nil {
		return nil
	}
	return unsafe.Slice(d.Items, int(d.Count))
}
func game3ListFirst(d *gui.ScrollListBoxData) int32 {
	if d == nil || d.Items == nil {
		return 0
	}
	items := game3ListItems(d)
	// Valid list state has populated<=capacity; malformed metadata is rejected
	// before a slice can escape the native item allocation.
	if int(d.Field_11_0) > len(items) {
		return 0
	}
	for i := 0; i < int(d.Field_11_0); i++ {
		if items[i].Field_0 > uint32(d.Field_13_1) {
			return int32(i)
		}
	}
	return 0
}

//export nox_xxx_wndListBoxFirstVisible
func nox_xxx_wndListBoxFirstVisible(w *C.nox_window) C.int {
	return C.int(game3ListFirst(game3ListData(w)))
}

//export nox_game3_first_data_go
func nox_game3_first_data_go(d *C.nox_scrollListBox_data) C.int {
	return C.int(game3ListFirst((*gui.ScrollListBoxData)(unsafe.Pointer(d))))
}

//export nox_xxx_wndListBox_4A2D10
func nox_xxx_wndListBox_4A2D10(w *C.nox_window, delta, update C.int) C.int {
	d := game3ListData(w)
	if d == nil || d.Items == nil || d.Field_11_0 > d.Count {
		return 0
	}
	first := game3ListFirst(d) + int32(delta)
	if first < 0 {
		first = 0
	} else if first >= int32(d.Field_11_0) {
		first = int32(d.Field_11_0) - 1
	}
	if update != 0 {
		if first <= 0 {
			d.Field_13_1 = 0
		} else {
			d.Field_13_1 = uint16(game3ListItems(d)[first-1].Field_0 + 1)
		}
	}
	slider := (*C.nox_window)(d.Field_9)
	if slider == nil || slider.widget_data == nil {
		return 0
	}
	s := (*game3SliderData)(slider.widget_data)
	s.Max = int32(d.Field_10) - int32(d.Field_13_0) + 3
	if s.Max < 0 {
		s.Max = 0
	}
	var thumb int32
	if slider.field_100 != nil {
		thumb = int32(slider.field_100.height)
	}
	s.Scale = 0
	if s.Max != 0 {
		s.Scale = float32(int32(slider.height)-thumb) / float32(s.Max)
	}
	if update != 0 {
		return C.int(game3WidgetEvent(slider, 16394, uintptr(int32(s.Max-int32(d.Field_13_1))), 0))
	}
	return 0
}

//export nox_xxx_wndListBox_4A3A70
func nox_xxx_wndListBox_4A3A70(w *C.nox_window) C.int {
	d := game3ListData(w)
	if d == nil || d.Items == nil || d.Field_11_0 > d.Count {
		return 0
	}
	var total uint32
	items := game3ListItems(d)
	for i := 0; i < int(d.Field_11_0); i++ {
		total += uint32(uint8(items[i].Field_130)) + 1
		items[i].Field_0 = total
	}
	d.Field_10 = total
	return nox_xxx_wndListBox_4A2D10(w, 0, 1)
}

func game3ListColor(i int32) uint32 {
	switch i {
	case 0:
		return uint32(C.nox_color_black_2650656)
	case 1:
		return *memmap.PtrUint32(0x852978, 4)
	case 2:
		return *memmap.PtrUint32(0x85B3FC, 956)
	case 3:
		return *memmap.PtrUint32(0x5D4594, 2597996)
	case 4:
		return uint32(C.nox_color_white_2523948)
	case 5:
		return uint32(C.nox_color_violet_2598268)
	case 6:
		return *memmap.PtrUint32(0x85B3FC, 940)
	case 7:
		return uint32(C.nox_color_red_2589776)
	case 8:
		return *memmap.PtrUint32(0x85B3FC, 984)
	case 9:
		return uint32(C.dword_8531A0_2572)
	case 10:
		return uint32(C.nox_color_green_2614268)
	case 11:
		return *memmap.PtrUint32(0x85B3FC, 944)
	case 12:
		return uint32(C.nox_color_cyan_2649820)
	case 13:
		return uint32(C.nox_color_blue_2650684)
	case 14:
		return uint32(C.nox_color_orange_2614256)
	case 15:
		return uint32(C.nox_color_yellow_2589772)
	case 16:
		return *memmap.PtrUint32(0x852978, 0)
	}
	return 0
}

//export nox_xxx_wndListBoxAddLine_4A3AC0
func nox_xxx_wndListBoxAddLine_4A3AC0(text *C.wchar2_t, color C.int, w *C.nox_window) C.int {
	d := game3ListData(w)
	if d == nil || d.Items == nil || d.Field_11_1 >= d.Count || d.Field_11_0 >= d.Count {
		return 0
	}
	item := &game3ListItems(d)[d.Field_11_1]
	item.Field_129 = uint32(w.draw_data.text_color)
	if color >= 0 && color < 17 {
		item.Field_129 = game3ListColor(int32(color))
	}
	if text == nil {
		item.Text[0] = 32
		item.Text[1] = 0
	} else {
		// Match wcsncpy padding: bytes beyond the first terminator are zeroed only
		// within the first255 code units. The final code unit is always terminated.
		p := (*uint16)(unsafe.Pointer(text))
		ended := false
		for i := 0; i < 255; i++ {
			var u uint16
			if !ended {
				u = *(*uint16)(unsafe.Add(unsafe.Pointer(p), uintptr(i)*2))
				ended = u == 0
			}
			item.Text[i] = u
		}
		item.Text[255] = 0
		n := 0
		for n < 255 && item.Text[n] != 0 {
			n++
		}
		if n > 0 && item.Text[n-1] == 10 {
			item.Text[n-1] = 0
		}
	}
	var height C.int
	if uint32(w.flags)&0x4000 != 0 {
		height = C.nox_xxx_guiFontHeightMB_43F320(w.draw_data.font)
	} else {
		width := int32(w.width) - 7
		if d.Field_3 != 0 {
			width = int32(w.width) - 17
		}
		C.nox_xxx_drawGetStringSize_43F840(w.draw_data.font, (*C.wchar2_t)(unsafe.Pointer(&item.Text[0])), nil, &height, C.int(width))
	}
	item.Field_130 = (item.Field_130 & 0xffffff00) | uint32(uint8(height))
	d.Field_11_1++
	d.Field_11_0++
	return nox_xxx_wndListBox_4A3A70(w)
}

//export sub_4A3090
func sub_4A3090(raw *C.short, index C.int) *C.short {
	d := (*gui.ScrollListBoxData)(unsafe.Pointer(raw))
	if d == nil || d.Field_12 == nil || index < 0 || int(index) >= int(d.Count) {
		return raw
	}
	// The native owner allocates exactly Count entries. The original extra
	// source word was copied into the last slot and immediately overwritten;
	// omit that dead read while preserving every resulting logical entry.
	a := unsafe.Slice(d.Field_12, int(d.Count))
	copy(a[int(index):int(d.Count)-1], a[int(index)+1:int(d.Count)])
	a[int(d.Count)-1] = ^uint32(0)
	return raw
}

// These four entrypoints already delegate to these native Go algorithms in
// the baseline. Preserve their extension/selection registry and callbacks.
//
//export nox_xxx_wndListboxProcWithoutData10_4A28E0
func nox_xxx_wndListboxProcWithoutData10_4A28E0(w *C.nox_window, e C.int, a, b C.uintptr_t) C.int {
	return C.int(guiEventRespC(gui.ScrollListBoxProc(asWindow(w), gui.AsWindowEvent(int(e), uintptr(a), uintptr(b)))))
}

//export nox_xxx_wndListboxProcWithData10_4A2DE0
func nox_xxx_wndListboxProcWithData10_4A2DE0(w *C.nox_window, e C.int, a, b C.uintptr_t) C.int {
	return C.int(guiEventRespC(gui.ScrollListBoxProc(asWindow(w), gui.AsWindowEvent(int(e), uintptr(a), uintptr(b)))))
}

//export nox_xxx_wndListboxProcPre_4A30D0
func nox_xxx_wndListboxProcPre_4A30D0(w *C.nox_window, e C.uint, a, b C.uintptr_t) C.uintptr_t {
	return guiEventRespC(gui.ScrollListBoxProcPre(asWindow(w), gui.AsWindowEvent(int(e), uintptr(a), uintptr(b))))
}

//export nox_xxx_wndListboxInit_4A3C00
func nox_xxx_wndListboxInit_4A3C00(w *C.nox_window, _ *C.nox_scrollListBox_data) {
	gui.ScrollListBoxInit(asWindow(w))
}

func game3ListDraw(w *C.nox_window, d *C.nox_window_data, image bool) C.int {
	if w == nil || d == nil {
		return 1
	}
	// Reuse the existing native renderer/selection owner and its independently
	// covered clip restoration. The named ABI entry selects its image variant.
	win := asWindow(w)
	flags := win.Flags
	defer func() { win.Flags = flags }()
	if image {
		win.Flags |= gui.StatusImage
	} else {
		win.Flags &^= gui.StatusImage
	}
	data := game3ListData(w)
	if data == nil || data.Field_11_0 > data.Count {
		return 1
	}
	items := game3ListItems(data)
	saved := make([]uint32, int(data.Field_11_0))
	for i := range saved {
		saved[i] = items[i].Field_130
		items[i].Field_130 = uint32(uint8(items[i].Field_130))
	}
	defer func() {
		for i, h := range saved {
			items[i].Field_130 = h
		}
	}()
	return C.int(gui.ScrollListBoxDraw(win, asWindowData(d)))
}

//export nox_xxx_wndListboxDrawNoImage_4A3C50
func nox_xxx_wndListboxDrawNoImage_4A3C50(w *C.nox_window, d *C.nox_window_data) C.int {
	return game3ListDraw(w, d, false)
}

//export nox_xxx_wndListboxDrawWithImage_4A3FC0
func nox_xxx_wndListboxDrawWithImage_4A3FC0(w *C.nox_window, d *C.nox_window_data) C.int {
	return game3ListDraw(w, d, true)
}
