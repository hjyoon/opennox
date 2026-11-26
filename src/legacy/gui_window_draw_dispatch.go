package legacy

/*
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3.h"
#include "GAME3_1.h"
#include "GAME3_2.h"
#include "client__gui__guibook.h"
#include "client__gui__guiinv.h"
#include "client__gui__guimeter.h"
#include "client__gui__guirank.h"
#include "client__gui__guispell.h"
#include "client__gui__guisumn.h"
#include "client__gui__servopts__access.h"
#include "client__gui__servopts__guiserv.h"

int sub_48C9F0(nox_window* win, void* draw);
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
)

type windowDrawProc func(*nox_window, *C.nox_window_data) C.int

func registerWindowDraw(addr unsafe.Pointer, proc windowDrawProc) {
	gui.RegisterWindowDrawC(addr, func(win *gui.Window, draw *gui.WindowData) int {
		return int(proc((*nox_window)(win.C()), (*C.nox_window_data)(draw.C())))
	})
}

func windowDrawWithOpaqueData(proc func(*nox_window, unsafe.Pointer) C.int) windowDrawProc {
	return func(win *nox_window, draw *C.nox_window_data) C.int {
		return proc(win, unsafe.Pointer(draw))
	}
}

func windowTooltipWithoutArgs(proc func() C.int) func(*gui.Window, *gui.WindowData, uintptr) {
	return func(*gui.Window, *gui.WindowData, uintptr) { proc() }
}

func windowTooltipWithIntPos(proc func(*nox_window, *C.nox_window_data, C.int) C.int) func(*gui.Window, *gui.WindowData, uintptr) {
	return func(win *gui.Window, draw *gui.WindowData, pos uintptr) {
		proc((*nox_window)(win.C()), (*C.nox_window_data)(draw.C()), C.int(pos))
	}
}

func windowTooltipWithNativePos(proc func(*nox_window, *C.nox_window_data, C.uintptr_t) C.int) func(*gui.Window, *gui.WindowData, uintptr) {
	return func(win *gui.Window, draw *gui.WindowData, pos uintptr) {
		proc((*nox_window)(win.C()), (*C.nox_window_data)(draw.C()), C.uintptr_t(pos))
	}
}

func init() {
	// Preserve the ABI-visible procedure address and use typed Go calls to
	// pass only the arguments the original draw or tooltip procedure expects.
	registerWindowDraw(C.nox_xxx_book_45BD30, nox_xxx_book_45BD30)
	registerWindowDraw(C.nox_xxx_drawHealthManaBar_471C00, nox_xxx_drawHealthManaBar_471C00)
	registerWindowDraw(C.nox_xxx_guiBottleSlotDrawFn_471A80, nox_xxx_guiBottleSlotDrawFn_471A80)
	registerWindowDraw(C.nox_xxx_guiDrawSummonBox_4C1FE0, nox_xxx_guiDrawSummonBox_4C1FE0)
	registerWindowDraw(C.nox_xxx_guiHealthManaTubeDraw_471D10, nox_xxx_guiHealthManaTubeDraw_471D10)
	registerWindowDraw(C.nox_xxx_inventoryDrawAllMB_463430, nox_xxx_inventoryDrawAllMB_463430)
	registerWindowDraw(C.nox_xxx_inventoryDrawProc_466580, nox_xxx_inventoryDrawProc_466580)
	registerWindowDraw(C.nox_xxx_quickBarDrawFn_45FBD0, nox_xxx_quickBarDrawFn_45FBD0)
	registerWindowDraw(C.nox_xxx_quickBarWarriorDraw_45FDE0, nox_xxx_quickBarWarriorDraw_45FDE0)
	registerWindowDraw(C.nox_xxx_quickbarDrawFn_460000, nox_xxx_quickbarDrawFn_460000)
	registerWindowDraw(C.nox_xxx_quickbarDraw_45FAC0, nox_xxx_quickbarDraw_45FAC0)
	registerWindowDraw(C.nox_xxx_quickbarTrapUpDownDraw_45F6F0, nox_xxx_quickbarTrapUpDownDraw_45F6F0)
	registerWindowDraw(C.nox_xxx_wndEditDrawNoImage_488160, nox_xxx_wndEditDrawNoImage_488160)
	registerWindowDraw(C.nox_xxx_wndEditDrawWithImage_488870, nox_xxx_wndEditDrawWithImage_488870)
	registerWindowDraw(C.nox_xxx_wndRadioButtonDrawNoImg_4A8820, nox_xxx_wndRadioButtonDrawNoImg_4A8820)
	registerWindowDraw(C.nox_xxx_wndRadioButtonDraw_4A8A20, nox_xxx_wndRadioButtonDraw_4A8A20)
	registerWindowDraw(C.sub_438C80, sub_438C80)
	registerWindowDraw(C.sub_438E30, sub_438E30)
	registerWindowDraw(C.sub_44E6F0, sub_44E6F0)
	registerWindowDraw(C.sub_44E8D0, sub_44E8D0)
	registerWindowDraw(C.sub_453350, sub_453350)
	registerWindowDraw(C.sub_453B80, sub_453B80)
	registerWindowDraw(C.sub_454A90, sub_454A90)
	registerWindowDraw(C.sub_455CD0, sub_455CD0)
	registerWindowDraw(C.sub_4560D0, sub_4560D0)
	registerWindowDraw(C.sub_456640, sub_456640)
	registerWindowDraw(C.sub_45F5D0, sub_45F5D0)
	registerWindowDraw(C.sub_45F9B0, sub_45F9B0)
	registerWindowDraw(C.sub_4625D0, sub_4625D0)
	registerWindowDraw(C.sub_466F50, sub_466F50)
	registerWindowDraw(C.sub_46F080, sub_46F080)
	registerWindowDraw(C.sub_470F40_draw, sub_470F40_draw)
	registerWindowDraw(C.sub_471250, sub_471250)
	registerWindowDraw(C.sub_471450, sub_471450)
	registerWindowDraw(C.sub_478970, sub_478970)
	registerWindowDraw(C.sub_479C40, sub_479C40)
	registerWindowDraw(C.sub_479CB0, sub_479CB0)
	registerWindowDraw(C.sub_4A49D0, sub_4A49D0)
	registerWindowDraw(C.sub_4A6D20, sub_4A6D20)
	registerWindowDraw(C.sub_4A6DC0, sub_4A6DC0)
	registerWindowDraw(C.sub_4AD570, sub_4AD570)
	registerWindowDraw(C.sub_4ADEF0, sub_4ADEF0)
	registerWindowDraw(C.sub_4C0D00, sub_4C0D00)
	registerWindowDraw(C.sub_4C24A0, sub_4C24A0)
	registerWindowDraw(C.sub_4C27F0, sub_4C27F0)
	registerWindowDraw(C.sub_4CEED0, sub_4CEED0)

	registerWindowDraw(C.nox_xxx_guiChatMode_4456E0, windowDrawWithOpaqueData(nox_xxx_guiChatMode_4456E0))
	registerWindowDraw(C.nox_xxx_movEax1Sub_4661C0, windowDrawWithOpaqueData(nox_xxx_movEax1Sub_4661C0))
	registerWindowDraw(C.sub_476E90, windowDrawWithOpaqueData(sub_476E90))
	registerWindowDraw(C.sub_48C9F0, windowDrawWithOpaqueData(sub_48C9F0))
	registerWindowDraw(C.sub_4AB420, windowDrawWithOpaqueData(sub_4AB420))
	registerWindowDraw(C.sub_4C0030, windowDrawWithOpaqueData(sub_4C0030))
	registerWindowDraw(C.sub_4C3410, windowDrawWithOpaqueData(sub_4C3410))

	registerWindowDraw(C.nox_xxx_bookDrawFn_45C7D0, func(win *nox_window, _ *C.nox_window_data) C.int {
		return nox_xxx_bookDrawFn_45C7D0((*C.uint32_t)(unsafe.Pointer(win)))
	})
	registerWindowDraw(C.nox_xxx_bookDrawIconFn_45CB30, func(win *nox_window, _ *C.nox_window_data) C.int {
		return nox_xxx_bookDrawIconFn_45CB30((*C.uint32_t)(unsafe.Pointer(win)))
	})
	registerWindowDraw(C.nox_xxx_bookDrawList_45BD40, func(win *nox_window, _ *C.nox_window_data) C.int {
		return nox_xxx_bookDrawList_45BD40(win)
	})
	registerWindowDraw(C.nox_xxx_quickbarButtonBookDraw_45EF30, func(_ *nox_window, _ *C.nox_window_data) C.int {
		return nox_xxx_quickbarButtonBookDraw_45EF30()
	})
	registerWindowDraw(C.nox_xxx_windowServerOptionsDrawProc_458500, func(win *nox_window, _ *C.nox_window_data) C.int {
		return nox_xxx_windowServerOptionsDrawProc_458500(win, 0)
	})
	registerWindowDraw(C.sub_45F8F0, func(_ *nox_window, _ *C.nox_window_data) C.int {
		return sub_45F8F0()
	})
	registerWindowDraw(C.sub_4C26F0, func(win *nox_window, draw *C.nox_window_data) C.int {
		return C.int(sub_4C26F0(win, draw))
	})

	gui.RegisterWindowTooltipC(C.nox_xxx_quickbarButtonBook_45F3F0, windowTooltipWithoutArgs(nox_xxx_quickbarButtonBook_45F3F0))
	gui.RegisterWindowTooltipC(C.sub_466160, windowTooltipWithoutArgs(sub_466160))
	gui.RegisterWindowTooltipC(C.sub_4661D0, windowTooltipWithoutArgs(sub_4661D0))
	gui.RegisterWindowTooltipC(C.sub_4710B0, windowTooltipWithoutArgs(sub_4710B0))
	gui.RegisterWindowTooltipC(C.sub_479D00, windowTooltipWithoutArgs(sub_479D00))

	gui.RegisterWindowTooltipC(C.sub_45F480, windowTooltipWithIntPos(sub_45F480))
	gui.RegisterWindowTooltipC(C.sub_4C1120, windowTooltipWithIntPos(sub_4C1120))

	gui.RegisterWindowTooltipC(C.sub_478E50, windowTooltipWithNativePos(sub_478E50))
	gui.RegisterWindowTooltipC(C.sub_4C2C20, windowTooltipWithNativePos(sub_4C2C20))
	gui.RegisterWindowTooltipC(C.sub_4C2CE0, windowTooltipWithNativePos(sub_4C2CE0))

	gui.RegisterWindowTooltipC(C.nox_xxx_book_45CF00, func(win *gui.Window, _ *gui.WindowData, _ uintptr) {
		nox_xxx_book_45CF00((*C.uint32_t)(win.C()))
	})
	gui.RegisterWindowTooltipC(C.nox_xxx_inventroryOnHovewerSub_4667E0, func(_ *gui.Window, _ *gui.WindowData, pos uintptr) {
		nox_xxx_inventroryOnHovewerSub_4667E0(0, 0, C.uint(pos))
	})
	gui.RegisterWindowTooltipC(C.nox_xxx_options_457AA0, func(_ *gui.Window, draw *gui.WindowData, _ uintptr) {
		nox_xxx_options_457AA0(0, (*C.uint8_t)(draw.C()))
	})
	gui.RegisterWindowTooltipC(C.nox_xxx_options_457B00, func(_ *gui.Window, draw *gui.WindowData, _ uintptr) {
		nox_xxx_options_457B00(0, (*C.uint8_t)(draw.C()))
	})
	gui.RegisterWindowTooltipC(C.sub_466620, func(_ *gui.Window, _ *gui.WindowData, pos uintptr) {
		sub_466620(0, 0, C.uint(pos))
	})
	gui.RegisterWindowTooltipC(C.sub_466E20, func(win *gui.Window, _ *gui.WindowData, _ uintptr) {
		sub_466E20((*C.uint32_t)(win.C()))
	})
}
