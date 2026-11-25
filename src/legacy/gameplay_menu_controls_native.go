package legacy

/*
#include "GAME1_3.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3_3.h"
#include "client__gui__guibook.h"
#include "client__gui__guiinv.h"
#include "client__gui__guiquit.h"

extern nox_window* dword_5d4594_1303452;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
)

//export sub_49B3E0
func sub_49B3E0() C.int {
	root := nox_new_window_from_file(internCStr("GGOver.wnd"), C.sub_49B420)
	C.dword_5d4594_1303452 = root
	if root == nil {
		return 0
	}
	nox_window_set_hidden(root, 1)
	nox_xxx_wnd_46ABB0_go(root, 0)
	return 1
}

//export sub_49B420
func sub_49B420(win *nox_window, event C.int, arg, value C.uintptr_t) C.int {
	if event != 16391 {
		return 0
	}
	id := nox_xxx_wndGetID_46B0A0((*nox_window)(unsafe.Pointer(uintptr(arg))))
	nox_xxx_clientPlaySoundSpecial_452D80(766, 100)
	switch id {
	case 10701:
		nox_client_quit_4460C0()
		sub_49B6B0()
	case 10702:
		packet := [...]byte{byte(netmsg.MSG_GAUNTLET), 3}
		nox_xxx_netClientSend2_4E53C0(31, C.nox_important_payload_t(unsafe.Pointer(&packet[0])), C.int(len(packet)), nil, 1)
		sub_49B6B0()
	}
	return 0
}

//export sub_49B490
func sub_49B490() C.int {
	result := nox_xxx_windowDestroyMB_46C4E0(C.dword_5d4594_1303452)
	C.dword_5d4594_1303452 = nil
	return C.int(result)
}

//export sub_49B6B0
func sub_49B6B0() C.int {
	win := C.dword_5d4594_1303452
	nox_window_set_hidden(win, 1)
	nox_xxx_wnd_46ABB0_go(win, 0)
	return C.int(nox_xxx_windowFocus_46B500(nil))
}

//export nox_xxx_consoleEsc_49B7A0
func nox_xxx_consoleEsc_49B7A0() {
	if nox_xxx_guiCursor_477600() != 0 || nox_video_inFadeTransition_44E0D0() != 0 {
		return
	}
	closed := sub_460660() != 0
	if sub_46A6A0() != 0 || closed {
		return
	}
	if sub_45D9B0() == 1 {
		sub_45D870()
		return
	}

	// Each handler runs before testing whether this Escape was consumed.
	closed = sub_4BFE40() != 0
	closed = nox_xxx_quickBarClose_4606B0() != 0 || closed
	closed = sub_462740() != 0 || closed
	if sub_44A4E0() != 0 || closed {
		return
	}
	switch sub_479590() {
	case 2, 3, 4:
		sub_4795A0(1)
		return
	}
	if sub_478040() != 0 || sub_479950() != 0 {
		return
	}

	closed = Sub_467C10() != 0
	closed = nox_xxx_bookHideMB_45ACA0(0) != 0 || closed
	closed = nox_gui_console_Hide_4512B0() != 0 || closed
	closed = sub_446780() != 0 || closed
	closed = nox_xxx_guiServerOptionsTryHide_4574D0() != 0 || closed
	closed = sub_48CAD0() != 0 || closed
	closed = sub_4AD9B0(1) != 0 || closed
	closed = sub_4C35B0(1) != 0 || closed
	if sub_46D6F0() != 0 || closed {
		return
	}
	if nox_xxx_game_4DCCB0() != 0 {
		sub_445C40()
	} else {
		nox_xxx_clientPlaySoundSpecial_452D80(231, 100)
	}
}
