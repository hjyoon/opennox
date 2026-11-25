package legacy

/*
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2_2.h"
#include "GAME3.h"
#include "GAME5_2.h"
#include "client__gui__window.h"
#include "client__shell__selchar.h"

extern nox_window* dword_5d4594_814984;
extern nox_window* dword_5d4594_815000;
extern nox_window* nox_wol_wnd_world_814980;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

//export sub_439D00
func sub_439D00(raw *C.int, event C.int, key C.uint, action C.int) C.int {
	win := (*nox_window)(unsafe.Pointer(raw))
	switch event {
	case 5:
		if nox_xxx_wndGetID_46B0A0(win) == 10020 && Get_nox_game_createOrJoin_815048() == 1 {
			sub_439D90(C.uint(uint16(key)), key>>16)
			return 1
		}
	case 21:
		switch key {
		case 1:
			if action == 2 {
				sub_4373A0()
			}
			return 1
		case 57:
			mouse := nox_client_getMousePos_4309F0()
			nox_window_call_field_93(win, 5, uintptr(C.int(mouse.x|mouse.y<<16)), 0)
		}
	}
	return 0
}

//export sub_439D90
func sub_439D90(x, y C.uint) C.int {
	result := sub_438DD0(x, y)
	if result == 0 {
		return result
	}
	region := uintptr(87528 + 8*Get_dword_587000_87412())
	*memmap.PtrUint16(0x5D4594, 814916) = uint16(x) + memmap.Uint16(0x587000, region) - 216
	*memmap.PtrUint16(0x5D4594, 814918) = uint16(y) + memmap.Uint16(0x587000, region+2) - 27
	sub_43B460()
	if sub_43BDB0()&2 != 0 {
		nox_xxx_setQuest_4D6F60(1)
	}
	hasSave := false
	if nox_xxx_isQuest_4D6F50() != 0 {
		hasSave = nox_client_countPlayerFiles04_4DC7D0() != 0
	} else {
		hasSave = nox_client_countPlayerFiles02_4DC630() != 0
	}
	if hasSave {
		C.sub_4A7A70(1)
		C.nox_game_showSelChar_4A4DB0()
		return C.nox_client_setCursorType_477610(0)
	}
	C.sub_4A7A70(0)
	C.nox_game_showSelClass_4A4840()
	C.nox_client_setCursorType_477610(0)
	return result
}

//export sub_43A920
func sub_43A920() C.int {
	nox_xxx_windowFocus_46B500(C.nox_wol_wnd_world_814980)
	if sub_43BE30() == 0 || memmap.Uint32(0x5D4594, 815084) == 0 {
		C.sub_44A400()
	}
	result := sub_43AF90(0)
	Set_dword_5d4594_815044(0)
	return result
}

//export nox_client_guiXxx_43A9D0
func nox_client_guiXxx_43A9D0() C.int {
	C.nox_xxx_wndClearCaptureMain_46ADE0(C.dword_5d4594_814984)
	sub_489FB0()
	sub_4A2890()
	if popup := C.dword_5d4594_815000; popup != nil && popup.parent == nil {
		nox_xxx_wnd_46C6E0(popup)
		nox_xxx_windowDestroyMB_46C4E0(C.dword_5d4594_815000)
		C.dword_5d4594_815000 = nil
	}
	if C.nox_wol_wnd_world_814980 != nil {
		nox_xxx_windowDestroyMB_46C4E0(C.nox_wol_wnd_world_814980)
		C.nox_wol_wnd_world_814980 = nil
	}
	sub_43A920()
	nox_xxx_windowFocus_46B500(nil)
	Set_nox_wol_server_result_cnt_815088(0)
	sub_49FFA0(0)
	C.sub_554D10()
	C.nox_client_setCursorType_477610(0)
	return C.sub_43DE40(nil)
}
