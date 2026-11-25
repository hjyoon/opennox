package legacy

/*
#include "client__gui__window.h"
#include "client__shell__noxworld.h"

extern nox_window* nox_wol_wnd_world_814980;
extern nox_window* nox_wol_wnd_gameList_815012;
extern nox_window *dword_5d4594_814984, *dword_5d4594_814988, *dword_5d4594_814992, *dword_5d4594_814996;
extern int dword_5d4594_815104;
*/
import "C"

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

func wolShowSelectedServer(point C.nox_point, server *C.nox_gui_server_ent_t) {
	Set_dword_5d4594_814624(unsafe.Pointer(server))
	position := C.int2{field_0: point.x, field_4: point.y}
	nox_client_gui_serverInfoBlockCheckExp_439370(&position, server)
}

func wolClickMapServer(id int32) {
	point := nox_client_getMousePos_4309F0()
	nearby := [2]C.uint32_t{C.uint32_t(point.x - 216), C.uint32_t(point.y - 27)}
	if sub_4A25C0(&nearby[0], sub_4A0020()) >= 2 {
		sub_4A2610(C.nox_wol_wnd_world_814980, &nearby[0], sub_4A0020())
	} else {
		wolShowSelectedServer(point, sub_4A0490(C.int(id-10070)))
	}
	nox_xxx_cursorSetTooltip_4776B0((*wchar2_t)(memmap.PtrOff(0x5D4594, 815112)))
}

func wolSelectWorldRegion(region int32) {
	Set_dword_587000_87412(int(region))
	sub_43A810()
	nox_client_refreshServerList_4378B0()
	if noxflags.HasGame(noxflags.GameFlag26) {
		point := (*[2]C.uint)(unsafe.Pointer(sub_4A7EF0()))
		sub_439D90(point[0]+216, point[1]+27)
	} else {
		sub_439D90(408, 239)
	}
}

func wolShowLegacyFilter() {
	Set_nox_game_createOrJoin_815048(0)
	nox_window_set_hidden(C.nox_wol_wnd_gameList_815012, 1)
	nox_window_set_hidden(C.dword_5d4594_814984, 1)
	nox_window_set_hidden(C.dword_5d4594_814988, 1)
	nox_window_set_hidden(C.dword_5d4594_814992, 0)
	sub_46AD20(C.nox_wol_wnd_world_814980, 10006, 10007, 0)
	sub_46ACE0(C.nox_wol_wnd_world_814980, 10047, 10051, 1)
	wolSetHeading(C.dword_5d4594_814996, wolWindowString("FilterMsg", 2186))
}

func wolStartLegacyHost(quest bool) {
	if !quest && noxflags.HasGame(noxflags.GameFlag25) {
		return
	}
	Set_nox_game_createOrJoin_815048(1)
	if quest {
		noxflags.UnsetGame(noxflags.GameNotQuest)
	} else {
		noxflags.SetGame(noxflags.GameNotQuest)
		nox_xxx_cliShowHideTubes_470AA0(0)
	}
	nox_client_setMousePos_430B10(408, 239)
	sub_4375C0(0)
	if wndIsShown_nox_xxx_wndIsShown_46ACC0(C.dword_5d4594_814992) == 0 {
		sub_489870()
	}
	if quest {
		nox_xxx_setQuest_4D6F60(1)
		nox_xxx_cliShowHideTubes_470AA0(1)
	} else if noxflags.HasGame(noxflags.GameFlag26) || sub_4D6F30() != 0 {
		return
	}
	Set_dword_587000_87412(0)
	child := nox_xxx_wndGetChildByID_46B0C0(C.nox_wol_wnd_world_814980, 10020)
	nox_window_call_field_93(child, 5, 15663512, 0)
}

func wolConfirmLegacyDialog() {
	switch sub_43AF80() {
	case 6:
		password := (*wchar2_t)(unsafe.Pointer(uintptr(sub_449E60(4))))
		var packet [11]wchar2_t
		nox_wcsncpy(&packet[2], password, 9)
		packet[10] = 0
		port := C.uint16_t(nox_client_getServerPort_43B320())
		address := int32(nox_client_getServerAddr_43B300())
		sub_5550D0(address, port, (*C.char)(unsafe.Pointer(&packet[0])))
		sub_43AF90(3)
		Set_qword_5d4594_814956(uint64(nox_platform_get_ticks() + 20000))
		sub_449EA0(0)
	case 10:
		sub_449E30(wolWindowString("Finding", 2209))
		sub_43AF90(11)
		sub_449EA0(0)
	case 1:
		sub_43A920()
		nox_game_showGameSel_4379F0()
	default:
		if C.dword_5d4594_815104 != 0 {
			Set_qword_5d4594_815068(uint64(nox_platform_get_ticks() + 1000))
		}
	}
}

func wolJoinLegacySelected() {
	if Get_dword_5d4594_815044() != 0 {
		return
	}
	if noxflags.GameFlag(sub_43B340()).Has(noxflags.GameModeQuest) {
		sub_4D6F80(1)
		nox_xxx_cliShowHideTubes_470AA0(1)
	}
	sub_43B460()
	var saved bool
	if sub_4D6F70() != 0 {
		saved = nox_client_countPlayerFiles04_4DC7D0() != 0
	} else {
		saved = nox_client_countPlayerFiles02_4DC630() != 0
	}
	if saved {
		sub_4A7A70(1)
		nox_game_showSelChar_4A4DB0()
	} else {
		sub_4A7A70(0)
		nox_game_showSelClass_4A4840()
	}
	Set_dword_5d4594_815044(1)
	sub_4A2890()
}

//export nox_xxx_windowMultiplayerSub_439E70
func nox_xxx_windowMultiplayerSub_439E70(_ *C.nox_window, event C.uint, arg, value C.uintptr_t) C.int {
	switch event {
	case 23:
		return 1
	case 16403, 16412:
		wolGameColumnsSend(int32(event), uintptr(arg), 0)
	case 0x4010:
		if nox_xxx_wndGetID_46B0A0((*nox_window)(unsafe.Pointer(uintptr(arg)))) == 10061 {
			point := nox_client_getMousePos_4309F0()
			wolShowSelectedServer(point, sub_4A28C0(C.int(value)))
		}
	case 0x4000:
		id := nox_xxx_wndGetID_46B0A0((*nox_window)(unsafe.Pointer(uintptr(arg))))
		if id == 10043 || id == 10044 {
			wolGameColumnsSend(int32(event), uintptr(arg), 0)
		}
	case 16391:
		id := nox_xxx_wndGetID_46B0A0((*nox_window)(unsafe.Pointer(uintptr(arg))))
		if id != 10043 && id != 10044 && id != 10035 && id != 10036 {
			nox_xxx_clientPlaySoundSpecial_452D80(766, 100)
		}
		switch {
		case id >= 10070:
			wolClickMapServer(id)
		case id >= 10054 && id <= 10057:
			wolSelectWorldRegion(id - 10054)
		case id >= 10047 && id <= 10051:
			sub_4379C0()
			nox_wol_servers_sortBtnHandler_4A0290(C.int(id))
			sub_4A0390()
		default:
			switch id {
			case 4001:
				wolConfirmLegacyDialog()
			case 4002:
				sub_43A920()
				nox_game_showGameSel_4379F0()
			case 10001:
				wolJoinLegacySelected()
			case 10002, 10003:
				wolStartLegacyHost(id == 10003)
			case 10004:
				Set_nox_game_createOrJoin_815048(0)
				sub_4383A0()
				nox_client_refreshServerList_4378B0()
			case 10005:
				wolShowLegacyFilter()
			case 10006:
				Set_nox_game_createOrJoin_815048(0)
				nox_client_refreshServerList_4378B0()
			case 10007:
				if Get_nox_game_createOrJoin_815048() != 0 {
					sub_4373A0()
				}
			case 10010:
				sub_4373A0()
			}
		}
	}
	return 0
}
