package legacy

/*
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2_2.h"
#include "client__gui__window.h"
#include "client__shell__noxworld.h"

*/
import "C"

import (
	"image"
	"unsafe"
)

//export sub_438C80
func sub_438C80(win *nox_window, draw *C.nox_window_data) C.int {
	mouse := nox_client_getMousePos_4309F0()
	popup := wolWindows.serverInfo
	if popup != nil && wndIsShown_nox_xxx_wndIsShown_46ACC0(popup) == 0 && Get_dword_5d4594_815044() == 0 {
		w := asWindow(popup)
		margin := image.Pt(32, 32)
		origin := w.GlobalPos()
		lo := origin.Sub(margin)
		hi := origin.Add(w.Size()).Add(margin)
		x, y := int(mouse.x), int(mouse.y)
		if x < lo.X || x > hi.X || y < lo.Y || y > hi.Y {
			nox_window_set_hidden(wolWindows.serverInfo, 1)
			nox_window_call_field_94_fnc(wolWindows.gameList, 16403, ^uintptr(0), 0, nil, 0)
			Set_dword_5d4594_815056(0)
			nox_xxx_wnd_46C6E0(wolWindows.serverInfo)
			nox_xxx_windowFocus_46B500(wolWindows.root)
		}
	}
	if sub_4A28B0() != 0 && !bool(nox_xxx_wndPointInWnd_46AAB0(game3NearbyWindow, mouse.x, mouse.y)) {
		sub_4A2890()
		nox_xxx_windowFocus_46B500(wolWindows.root)
	}
	if Get_nox_game_createOrJoin_815048() != 0 && sub_438DD0(C.uint(mouse.x), C.uint(mouse.y)) != 0 {
		C.nox_client_setCursorType_477610(9)
	} else if C.sub_44A4A0() == 0 {
		C.nox_client_setCursorType_477610(0)
	}
	return 1
}

//export sub_438DD0
func sub_438DD0(x, y C.uint) C.int {
	if Get_dword_587000_87412() == -1 {
		return C.int(bool2int(x > 216 && x < 600 && y > 27 && y < 451))
	}
	return C.int(bool2int(x > 226 && x < 590 && y > 37 && y < 441))
}

func wolGameColumnsSend(event int32, arg, arg2 uintptr) {
	for _, column := range wolWindows.gameColumns {
		nox_window_call_field_94_fnc(column, event, arg, arg2, nil, 0)
	}
}

//export sub_438EF0
func sub_438EF0(win *nox_window, event C.int, arg, arg2 C.uintptr_t) C.int {
	if event != 19 && event != 20 {
		return nox_xxx_wndListboxProcWithoutData10_4A28E0(win, event, arg, arg2)
	}
	data := game3ListData(win)
	if data == nil {
		return 0
	}
	control := uintptr(data.Field_7)
	if event == 20 {
		control = uintptr(data.Field_8)
	}
	nox_window_call_field_94_fnc(wolWindows.gameList, 16391, control, 0, nil, 0)
	wolGameColumnsSend(16391, control, 0)
	return 0
}

//export sub_439050
func sub_439050(win *nox_window, event C.uint, arg, arg2 C.uintptr_t) C.int {
	switch event {
	case 23:
		return 1
	case 0x4000:
		sender := (*nox_window)(unsafe.Pointer(uintptr(arg)))
		list := wolWindows.gameList
		if sender == nox_xxx_wndGetChildByID_46B0C0(list, 10043) ||
			sender == nox_xxx_wndGetChildByID_46B0C0(list, 10044) {
			wolGameColumnsSend(0x4000, uintptr(arg), 0)
		}
	case 0x4009:
		nox_xxx_wndListboxProcPre_4A30D0(win, event, arg, arg2)
		wolGameColumnsSend(16412, uintptr(game3ListFirst(game3ListData(win))), 0)
	case 0x400e, 0x400f:
		wolGameColumnsSend(int32(event), uintptr(arg), uintptr(arg2))
	case 16400:
		id := nox_xxx_wndGetID_46B0A0((*nox_window)(unsafe.Pointer(uintptr(arg))))
		if id >= 10038 && id <= 10042 {
			wolGameColumnsSend(16403, uintptr(arg2), 0)
			if uintptr(arg2) < uintptr(Get_nox_wol_server_result_cnt_815088()) {
				pos := nox_client_getMousePos_4309F0()
				record := sub_4A04C0(C.int(arg2))
				Set_dword_5d4594_814624(unsafe.Pointer(record))
				C.nox_client_gui_serverInfoBlockCheckExp_439370((*C.int2)(unsafe.Pointer(&pos)), record)
			}
		}
	case 16403, 16412:
		wolGameColumnsSend(int32(event), uintptr(arg), 0)
	}
	return C.int(nox_xxx_wndListboxProcPre_4A30D0(win, event, arg, arg2))
}

//export sub_439450
func sub_439450(x, y C.int, dst *C.uint32_t) *C.uint32_t {
	if dst == nil {
		return nil
	}
	out := unsafe.Slice(dst, 2)
	x -= 65
	y -= 20
	if x+130 > 600 {
		x = 470
	}
	if y+120 > 451 {
		y = 331
	}
	minY := C.uint32_t(27)
	if Get_dword_587000_87408() == 1 {
		minY = 55
	}
	out[0] = C.uint32_t(max(x, 216))
	out[1] = max(C.uint32_t(y), minY)
	return dst
}
