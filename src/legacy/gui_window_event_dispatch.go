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
#include "client__gui__guiinput.h"
#include "client__gui__guiinv.h"
#include "client__gui__guimeter.h"
#include "client__gui__guiquit.h"
#include "client__gui__servopts__access.h"
#include "client__gui__servopts__guiserv.h"
#include "client__gui__servopts__playrlst.h"
#include "client__gui__servopts__spelllst.h"
#include "client__shell__inputcfg__inputcfg.h"
#include "client__shell__noxworld.h"
#include "client__shell__optsback.h"
#include "client__shell__selclass.h"
*/
import "C"

import (
	"slices"
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
)

type windowEventProc func(*nox_window, C.int, C.uintptr_t, C.uintptr_t) C.int

func registerWindowEvent(addr unsafe.Pointer, proc windowEventProc) {
	gui.RegisterWindowProcC(addr, func(win *gui.Window, event int, a, b uintptr) uintptr {
		r := proc((*nox_window)(win.C()), C.int(event), C.uintptr_t(a), C.uintptr_t(b))
		return uintptr(uint32(r))
	})
}

func windowEventControl(event C.int, arg C.uintptr_t, controlEvents ...C.int) *nox_window {
	// The first payload also carries scalars for mouse and control events.
	if !slices.Contains(controlEvents, event) {
		return nil
	}
	return (*nox_window)(unsafe.Pointer(uintptr(arg)))
}

func windowControlProc(proc func(*nox_window, C.int, *nox_window, C.uintptr_t) C.int, controlEvents ...C.int) windowEventProc {
	return func(win *nox_window, event C.int, control, arg C.uintptr_t) C.int {
		return proc(win, event, windowEventControl(event, control, controlEvents...), arg)
	}
}

func windowUnsignedProc(proc func(*nox_window, C.uint, C.uintptr_t, C.uintptr_t) C.int) windowEventProc {
	return func(win *nox_window, event C.int, a, b C.uintptr_t) C.int {
		return proc(win, C.uint(event), a, b)
	}
}

func windowEventOnlyProc(proc func(*nox_window, C.int) C.int) windowEventProc {
	return func(win *nox_window, event C.int, _, _ C.uintptr_t) C.int {
		return proc(win, event)
	}
}

func windowUnsignedEventOnlyProc(proc func(*nox_window, C.uint) C.int) windowEventProc {
	return func(win *nox_window, event C.int, _, _ C.uintptr_t) C.int {
		return proc(win, C.uint(event))
	}
}

func registerWindowInitializer(addr unsafe.Pointer, proc func(*nox_window, C.int, C.uintptr_t, C.uintptr_t) C.uintptr_t) {
	gui.RegisterWindowInitializerC(addr, func(win *gui.Window, event int, a, b uintptr) uintptr {
		return uintptr(proc((*nox_window)(win.C()), C.int(event), C.uintptr_t(a), C.uintptr_t(b)))
	})
}

func init() {
	// C keeps the exported procedure addresses in the native window layout.
	// Dispatch the Go implementations with their declared argument types.
	registerWindowEvent(C.nox_client_wndQuestBriefProc_44E630, nox_client_wndQuestBriefProc_44E630)
	registerWindowEvent(C.nox_xxx_Proc_46F070, nox_xxx_Proc_46F070)
	registerWindowEvent(C.nox_xxx_XorEaxEaxSub_464BA0, nox_xxx_XorEaxEaxSub_464BA0)
	registerWindowEvent(C.nox_xxx_clientOrderCreature_4C2A60, nox_xxx_clientOrderCreature_4C2A60)
	registerWindowEvent(C.nox_xxx_guiServerOptionsProc_458590, nox_xxx_guiServerOptionsProc_458590)
	registerWindowEvent(C.nox_xxx_inventoryWndProc_464BB0, nox_xxx_inventoryWndProc_464BB0)
	registerWindowEvent(C.nox_xxx_quickbar_45F8D0, nox_xxx_quickbar_45F8D0)
	registerWindowEvent(C.nox_xxx_windowAccessProc_454BA0, nox_xxx_windowAccessProc_454BA0)
	registerWindowEvent(C.nox_xxx_windowMplayFilterProc_489E70, nox_xxx_windowMplayFilterProc_489E70)
	registerWindowEvent(C.nox_xxx_wndEditProc_487D70, nox_xxx_wndEditProc_487D70)
	registerWindowEvent(C.nox_xxx_wndRadioButtonProc_4A84E0, nox_xxx_wndRadioButtonProc_4A84E0)
	registerWindowEvent(C.nox_xxx_wndSummonBigButtonProc_4C24B0, nox_xxx_wndSummonBigButtonProc_4C24B0)
	registerWindowEvent(C.nox_xxx_wndSummonProc_4C2B10, nox_xxx_wndSummonProc_4C2B10)
	registerWindowEvent(C.sub_438EF0, sub_438EF0)
	registerWindowEvent(C.sub_464770, sub_464770)
	registerWindowEvent(C.sub_464BD0, sub_464BD0)
	registerWindowEvent(C.sub_466550, sub_466550)
	registerWindowEvent(C.sub_466BA0, sub_466BA0)
	registerWindowEvent(C.sub_466BF0, sub_466BF0)
	registerWindowEvent(C.sub_46A7E0, sub_46A7E0)
	registerWindowEvent(C.sub_46F060, sub_46F060)
	registerWindowEvent(C.sub_478480, sub_478480)
	registerWindowEvent(C.sub_478650, sub_478650)
	registerWindowEvent(C.sub_479BE0, sub_479BE0)
	registerWindowEvent(C.sub_49B420, sub_49B420)
	registerWindowEvent(C.sub_4A7270, sub_4A7270)
	registerWindowEvent(C.sub_4BE320, sub_4BE320)
	registerWindowEvent(C.sub_4BE330, sub_4BE330)
	registerWindowEvent(C.sub_4BFDD0, sub_4BFDD0)
	registerWindowEvent(C.sub_4C0630, sub_4C0630)
	registerWindowEvent(C.sub_4C0C90, sub_4C0C90)
	registerWindowEvent(C.sub_4C2BD0, sub_4C2BD0)
	registerWindowEvent(C.sub_4C2BE0, sub_4C2BE0)
	registerWindowEvent(C.sub_4C3A60, sub_4C3A60)
	registerWindowEvent(C.sub_4C3A90, sub_4C3A90)
	registerWindowEvent(C.sub_4C3CD0, sub_4C3CD0)
	registerWindowEvent(C.sub_4C3EB0, sub_4C3EB0)
	registerWindowEvent(C.sub_4CBE70, sub_4CBE70)
	registerWindowEvent(C.sub_4CBF60, sub_4CBF60)
	registerWindowEvent(C.sub_4CC140, sub_4CC140)
	registerWindowEvent(C.sub_4CC170, sub_4CC170)

	registerWindowEvent(C.nox_xxx_guiKick_48D0A0, windowControlProc(nox_xxx_guiKick_48D0A0, 0x4007))
	registerWindowEvent(C.nox_xxx_menuGameOnButton_445840, windowControlProc(nox_xxx_menuGameOnButton_445840, 0x4007))
	registerWindowEvent(C.nox_xxx_windowAdvancedServProc_4BDDB0, windowControlProc(nox_xxx_windowAdvancedServProc_4BDDB0, 0x4007))
	registerWindowEvent(C.nox_xxx_windowOptionsProc_4ADF30, windowControlProc(nox_xxx_windowOptionsProc_4ADF30, 0x4007, 0x4009, 0x400c))
	registerWindowEvent(C.nox_xxx_windowServerOptionsGeneralProc_4AD5D0, windowControlProc(nox_xxx_windowServerOptionsGeneralProc_4AD5D0, 0x4007, 0x4010))
	registerWindowEvent(C.nox_xxx_wnd_49C760, windowControlProc(nox_xxx_wnd_49C760, 0x4007))
	registerWindowEvent(C.sub_4533D0, windowControlProc(sub_4533D0, 0x4000, 0x4007))
	registerWindowEvent(C.sub_453C00, windowControlProc(sub_453C00, 0x4000, 0x4007))
	registerWindowEvent(C.sub_4567C0, windowControlProc(sub_4567C0, 0x4007, 0x4010))
	registerWindowEvent(C.sub_49CA60, windowControlProc(sub_49CA60, 0x4007))
	registerWindowEvent(C.sub_4AABE0, windowControlProc(sub_4AABE0, 0x4007, 0x4009, 0x400c))
	registerWindowEvent(C.sub_4BDDA0, windowControlProc(sub_4BDDA0))
	registerWindowEvent(C.sub_4BFCD0, windowControlProc(sub_4BFCD0, 0x4007))
	registerWindowEvent(C.sub_4C01C0, windowControlProc(sub_4C01C0, 0x4007))

	registerWindowEvent(C.nox_xxx_guiServerOptionsProcPre_4585D0, windowUnsignedProc(nox_xxx_guiServerOptionsProcPre_4585D0))
	registerWindowEvent(C.nox_xxx_quickbarTrapProc_45FB90, windowUnsignedProc(nox_xxx_quickbarTrapProc_45FB90))
	registerWindowEvent(C.nox_xxx_windowMultiplayerSub_439E70, windowUnsignedProc(nox_xxx_windowMultiplayerSub_439E70))
	registerWindowEvent(C.sub_439050, windowUnsignedProc(sub_439050))

	registerWindowEvent(C.nox_xxx_bookWndProc_45B070, windowEventOnlyProc(nox_xxx_bookWndProc_45B070))
	registerWindowEvent(C.nox_xxx_book_45B210, windowEventOnlyProc(nox_xxx_book_45B210))
	registerWindowEvent(C.nox_xxx_guiBottleSlotProc_471B90, windowEventOnlyProc(nox_xxx_guiBottleSlotProc_471B90))
	registerWindowEvent(C.nox_xxx_guiHealthManaTubeProc_472100, windowEventOnlyProc(nox_xxx_guiHealthManaTubeProc_472100))
	registerWindowEvent(C.sub_470E90, windowEventOnlyProc(sub_470E90))

	registerWindowEvent(C.nox_xxx_quickbarButtonBookWnd_45F450, windowUnsignedEventOnlyProc(nox_xxx_quickbarButtonBookWnd_45F450))
	registerWindowEvent(C.nox_xxx_quickbarTrapUpDownProc_45F630, windowUnsignedEventOnlyProc(nox_xxx_quickbarTrapUpDownProc_45F630))
	registerWindowEvent(C.sub_45F520, windowUnsignedEventOnlyProc(sub_45F520))

	registerWindowEvent(C.nox_xxx_bookChildWndProcMB_45B360, func(win *nox_window, event C.int, _, _ C.uintptr_t) C.int {
		return nox_xxx_bookChildWndProcMB_45B360((*C.uint32_t)(unsafe.Pointer(win)), C.uint(event))
	})
	registerWindowEvent(C.nox_xxx_bookListWndProc_45B5F0, func(win *nox_window, event C.int, position, _ C.uintptr_t) C.int {
		return nox_xxx_bookListWndProc_45B5F0(win, C.uint(event), C.uint(position))
	})
	registerWindowEvent(C.nox_xxx_bookWndFn_45CC10, func(win *nox_window, event C.int, position, _ C.uintptr_t) C.int {
		return nox_xxx_bookWndFn_45CC10((*C.uint32_t)(unsafe.Pointer(win)), event, C.uint(position))
	})
	registerWindowEvent(C.nox_xxx_guiDialog_479B00, func(_ *nox_window, event C.int, control, _ C.uintptr_t) C.int {
		button := (*int32)(unsafe.Pointer(windowEventControl(event, control, 0x4007)))
		return C.int(nox_xxx_guiDialog_479B00(0, int32(event), button, 0))
	})
	registerWindowEvent(C.nox_xxx_quickBarWnd_45EF50, func(win *nox_window, event C.int, position, _ C.uintptr_t) C.int {
		return nox_xxx_quickBarWnd_45EF50(win, event, C.uint(position))
	})
	registerWindowEvent(C.nox_xxx_quickbarTrapButtonProc_45F7A0, func(win *nox_window, event C.int, position, _ C.uintptr_t) C.int {
		return nox_xxx_quickbarTrapButtonProc_45F7A0(win, C.uint(event), C.uint(position))
	})
	registerWindowEvent(C.sub_439D00, func(win *nox_window, event C.int, key, action C.uintptr_t) C.int {
		return sub_439D00((*C.int)(unsafe.Pointer(win)), event, C.uint(key), C.int(action))
	})
	registerWindowEvent(C.sub_4466C0, func(_ *nox_window, event C.int, _, _ C.uintptr_t) C.int {
		return sub_4466C0(0, event, 0, 0)
	})
	registerWindowEvent(C.sub_45EF40, func(_ *nox_window, _ C.int, _, _ C.uintptr_t) C.int {
		return sub_45EF40()
	})
	registerWindowEvent(C.sub_46A820, func(win *nox_window, event C.int, a, b C.uintptr_t) C.int {
		return sub_46A820(win, event, C.int(a), C.int(b))
	})
	registerWindowEvent(C.sub_4A4A20, func(win *nox_window, event C.int, control, arg C.uintptr_t) C.int {
		return sub_4A4A20(win, event, windowEventControl(event, control, 0x4005, 0x4007), C.int(arg))
	})
	registerWindowEvent(C.sub_4AB340, func(win *nox_window, event C.int, a, b C.uintptr_t) C.int {
		return sub_4AB340(win, event, C.int(a), C.int(b))
	})
	registerWindowEvent(C.sub_4AB390, func(win *nox_window, event C.int, control, arg C.uintptr_t) C.int {
		return sub_4AB390(win, event, windowEventControl(event, control, 0x4007), C.int(arg))
	})
	registerWindowEvent(C.sub_4CF060, func(win *nox_window, event C.int, control, arg C.uintptr_t) C.int {
		// The rule list also uses the literal 1 as a control sentinel.
		if control == 1 {
			if event == 0x4003 {
				return 0
			}
			return 1
		}
		return sub_4CF060(win, C.uint(event), windowEventControl(event, control, 0x4003, 0x4007, 0x4010), arg)
	})

	gui.RegisterWindowProcC(C.nox_xxx_wndEditProcPre_488710, func(win *gui.Window, event int, a, b uintptr) uintptr {
		return uintptr(nox_xxx_wndEditProcPre_488710((*nox_window)(unsafe.Pointer(win)), C.uint(event), C.uintptr_t(a), C.uintptr_t(b)))
	})
	gui.RegisterWindowProcC(C.sub_4A18E0, func(win *gui.Window, event int, a, b uintptr) uintptr {
		return uintptr(sub_4A18E0(win.C(), int32(event), a, b))
	})
	gui.RegisterWindowProcC(C.sub_4A7330, func(win *gui.Window, event int, a, b uintptr) uintptr {
		return uintptr(sub_4A7330((*nox_window)(win.C()), C.int(event), C.uintptr_t(a), C.uintptr_t(b)))
	})
	registerWindowInitializer(C.nox_xxx_wndProc_44E6E0, nox_xxx_wndProc_44E6E0)
	registerWindowInitializer(C.sub_466220, sub_466220)
}
