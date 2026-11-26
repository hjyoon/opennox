package legacy

/*
#include "GAME2_1.h"
#include "GAME2_3.h"

extern void* dword_5d4594_1096288;
extern uint64_t qword_581450_9512, qword_581450_9544;
extern uint32_t dword_5d4594_1096264;
extern int nox_win_width, nox_win_height;
*/
import "C"

import (
	"math"
	"strconv"
	"unsafe"

	"github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

func meterWindowPosition(win *C.nox_window) (C.int, C.int) {
	var x, y C.uint
	nox_client_wndGetPosition_46AA60(win, &x, &y)
	return C.int(x), C.int(y)
}

//export sub_470F40_draw
func sub_470F40_draw(win *C.nox_window, _ *C.nox_window_data) C.int {
	meter := &clientMeterRecords[uintptr(win.widget_data)]
	x, y := meterWindowPosition(win)
	var w, h C.int
	nox_window_get_size(win, &w, &h)
	meter.color_2 = meter.color_1
	if meter.field_2 == 0 {
		sub_465D50_draw(win)
		return 1
	}
	progress := int32((meter.field_1 << 8) / meter.field_2)
	dial := true
	if progress >= 256 {
		dial = false
		progress = 1
		if weapon := sub_4678D0(); weapon != nil {
			current, maximum := float64(weapon.field_73_1), float64(weapon.field_73_2)
			if current < maximum*memmap.Float64(0x581450, 9608) {
				meter.color_2 = C.uint32_t(memmap.Uint32(0x85B3FC, 940))
				dial = true
			} else if current < maximum*math.Float64frombits(uint64(C.qword_581450_9544)) {
				meter.color_2 = renderPalette.yellow
				dial = true
			}
		}
	}
	if dial {
		nox_client_drawEnableAlpha_434560(1)
		nox_client_drawSetAlpha_434580(0x40)
		sub_4AE6F0(int32(x+w/2), int32(y+h/2), int32(w/2), progress, int32(meter.color_2))
		nox_client_drawEnableAlpha_434560(0)
	}
	sub_465D50_draw(win)
	return 1
}

const meterChargeRows = 61

func drawMeterChargeRow(x, y C.int, row uintptr) {
	pos := memmap.PtrT[[3]byte](0x587000, 147904+8*row)
	C.nox_client_drawRectFilledOpaque_49CE30(x+C.int(pos[0]), y+C.int(pos[1]), C.int(pos[2]), 1)
}

//export sub_471250
func sub_471250(win *C.nox_window, _ *C.nox_window_data) C.int {
	meter := &clientMeterRecords[uintptr(win.widget_data)]
	x, y := meterWindowPosition(win)
	if meter.field_2 == 0 {
		nox_client_drawSetColor_434460(int32(meter.color_1))
		nox_client_drawEnableAlpha_434560(1)
		for row := range uintptr(meterChargeRows) {
			drawMeterChargeRow(x, y, row)
		}
		nox_client_drawEnableAlpha_434560(0)
		return 1
	}

	sequence := uint32(nox_xxx_bookGet_430B40_get_mouse_prev_seq())
	maximum := int32(meter.field_2)
	if maximum <= 0 {
		return 1
	}
	gap := int32(1)
	if maximum > 30 {
		gap = 0
	}
	rowsPerCharge := (gap+meterChargeRows)/maximum - gap
	fraction := float32(float64(gap+meterChargeRows-maximum*((gap+meterChargeRows)/maximum)) / float64(maximum))
	phase := float32(0.001)
	unit := math.Float64frombits(uint64(C.qword_581450_9512))
	bottom := meterChargeRows - rowsPerCharge
	for cell := uint32(0); ; cell++ {
		color := meter.color_2
		if cell < uint32(meter.field_1) {
			color = meter.color_1
		}
		row, count := bottom, max(1, rowsPerCharge)
		bottom -= rowsPerCharge + gap
		phase = float32(phase + fraction)
		if float64(phase) >= unit {
			bottom--
			row--
			count++
			phase = float32(float64(phase) - unit)
		}
		nox_client_drawSetColor_434460(int32(color))
		nox_client_drawEnableAlpha_434560(1)
		if row < 0 {
			count += row
			row = 0
		}
		for end := min(row+count, meterChargeRows); row < end; row++ {
			last := memmap.PtrUint32(0x587000, 147908+8*uintptr(row))
			if *last != sequence {
				drawMeterChargeRow(x, y, uintptr(row))
				*last = sequence
			}
		}
		nox_client_drawEnableAlpha_434560(0)
		if cell+1 >= uint32(meter.field_2) {
			break
		}
	}
	return 1
}

func meterDecimalText(buffer *[16]uint16, number int32) *wchar2_t {
	p := (*wchar2_t)(unsafe.Pointer(buffer))
	WStrCopy(p, len(buffer), strconv.FormatInt(int64(number), 10))
	return p
}

//export sub_471450
func sub_471450(win *C.nox_window, _ *C.nox_window_data) C.int {
	var buffer [16]uint16
	meter := &clientMeterRecords[uintptr(win.widget_data)]
	text := meterDecimalText(&buffer, int32(meter.field_1))
	nox_xxx_drawSetTextColor_434390(int32(renderPalette.white))
	var width C.int
	nox_xxx_drawGetStringSize_43F840(C.dword_5d4594_1096288, text, &width, nil, 0)
	x, y := meterWindowPosition(win)
	nox_xxx_drawString_43F6E0(C.dword_5d4594_1096288, text, int32(x-width/2+8), int32(y+1))
	return 1
}

//export nox_xxx_guiBottleSlotDrawFn_471A80
func nox_xxx_guiBottleSlotDrawFn_471A80(win *C.nox_window, _ *C.nox_window_data) C.int {
	index := uintptr(win.widget_data)
	offset := healthQuickItem + 536*index
	x, y := meterWindowPosition(win)
	if memmap.Uint16(0x5D4594, offset+4) != 0 {
		drawable := &clientPotionDrawables[index]
		if drawable.draw_func != nil {
			drawable.pos.x, drawable.pos.y = C.intptr_t(x+14), C.intptr_t(y+15)
			viewport := legacyViewportFromWords32(memmap.PtrT[[13]uint32](0x5D4594, 1091908))
			ccall.CallIntPtr2(unsafe.Pointer(drawable.draw_func), viewport.C(), unsafe.Pointer(drawable))
		}
		var buffer [16]uint16
		nox_xxx_drawSetTextColor_434390(int32(renderPalette.white))
		count := memmap.Uint16(0x5D4594, offset+4)
		text := meterDecimalText(&buffer, int32(count))
		height := nox_xxx_guiFontHeightMB_43F320(C.dword_5d4594_1096288)
		nox_xxx_drawString_43F6E0(C.dword_5d4594_1096288, text, int32(x-2), int32(y)-height+10)
	}
	label := (*wchar2_t)(memmap.PtrOff(0x5D4594, offset-8))
	height := nox_xxx_guiFontHeightMB_43F320(C.dword_5d4594_1096288)
	nox_xxx_drawString_43F6E0(C.dword_5d4594_1096288, label, int32(x-2), int32(y)-height+33)
	return 1
}

//export nox_xxx_guiBottleSlotProc_471B90
func nox_xxx_guiBottleSlotProc_471B90(win *C.nox_window, event C.int) C.int {
	switch event {
	case 5:
		offset := healthQuickItem + 536*uintptr(win.widget_data)
		if id := memmap.Int32(0x5D4594, offset); id != 0 {
			nox_xxx_cliUseCurePoison_4674E0(C.int(id))
		}
		return 1
	case 8, 12, 16:
		return 0
	default:
		return 1
	}
}

//export nox_xxx_drawHealthManaBar_471C00
func nox_xxx_drawHealthManaBar_471C00(win *C.nox_window, _ *C.nox_window_data) C.int {
	index := uintptr(win.widget_data)
	meter := &clientMeterRecords[index]
	if nox_xxx_clientIsObserver_4372E0() != 0 || Get_nox_gameDisableMapDraw_5d4594_2650672() != 0 ||
		noxflags.HasGame(noxflags.GameFlag21|noxflags.GameFlag24) {
		return 1
	}
	x, y := C.nox_win_width/2+15, C.nox_win_height/2-48
	if index != 0 {
		x += 6
	}
	var fill C.int
	if meter.field_2 != 0 {
		fill = C.int(48 * meter.field_1 / meter.field_2)
	}
	nox_client_drawSetColor_434460(int32(renderPalette.black))
	C.nox_client_drawRectFilledOpaque_49CE30(x, y, 2, 48)
	nox_client_drawSetColor_434460(int32(meter.color_1))
	C.nox_client_drawRectFilledOpaque_49CE30(x, y-fill+48, 2, fill)
	border := int32(renderPalette.violet)
	if index != 0 {
		border = memmap.Int32(0x85B3FC, 944)
	} else if C.dword_5d4594_1096264 != 0 {
		border = memmap.Int32(0x85B3FC, 984)
	}
	nox_client_drawSetColor_434460(border)
	C.nox_client_drawBorderLines_49CC70(x-1, y-1, 4, 50)
	return 1
}
