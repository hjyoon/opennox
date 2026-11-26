package legacy

/*
#include "GAME1_2.h"
#include "client__gui__window.h"
#include "client__shell__noxworld.h"

extern uint32_t dword_587000_87408, dword_5d4594_815096, dword_5d4594_815100;
*/
import "C"

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

var (
	wolConnectDeadline uint64
	wolRefreshDeadline uint64
	wolConnectionError uint32
)

func wolWindowString(id string, line int32) *wchar2_t {
	const file = `C:\NoxPost\src\client\shell\noxworld.c`
	return nox_strman_loadString_40F1D0(internCStr(id), nil, internCStr(file), line)
}

func wolSetHeading(win *nox_window, text *wchar2_t) {
	nox_window_call_field_94_fnc(win, 16385, uintptr(unsafe.Pointer(text)), 0, nil, 0)
}

//export sub_4383A0
func sub_4383A0() C.int {
	if wndIsShown_nox_xxx_wndIsShown_46ACC0(wolWindows.filter) == 0 {
		nox_window_set_hidden(wolWindows.filter, 1)
		sub_489870()
	}
	nox_window_set_hidden(wolWindows.gameList, 0)
	nox_window_set_hidden(wolWindows.mapPanel, 1)
	nox_window_set_hidden(wolWindows.regionPanel, 1)
	sub_46AD20(wolWindows.root, 10006, 10007, 1)
	sub_46ACE0(wolWindows.root, 10047, 10051, 0)
	wolSetHeading(wolWindows.heading, wolWindowString("ListJoinServer", 418))
	C.dword_587000_87408 = 1
	return 0
}

//export sub_438770
func sub_438770() C.int {
	switch sub_43AF80() {
	case 0:
		listVisible := wndIsShown_nox_xxx_wndIsShown_46ACC0(wolWindows.mapPanel) == 0 ||
			wndIsShown_nox_xxx_wndIsShown_46ACC0(wolWindows.regionPanel) == 0 ||
			wndIsShown_nox_xxx_wndIsShown_46ACC0(wolWindows.gameList) == 0
		if listVisible && Get_nox_game_createOrJoin_815048() == 0 && Get_dword_5d4594_815044() == 0 &&
			Sub_43AF30() == 0 && wndIsShown_nox_xxx_wndIsShown_46ACC0(wolWindows.serverInfo) != 0 {
			if uint64(nox_platform_get_ticks()) > wolRefreshDeadline {
				nox_client_refreshServerList_4378B0()
			} else {
				sub_438770_waitList()
			}
		}
	case 2:
		sub_438BD0()
		sub_43AF90(1)
	case 3:
		if uint64(nox_platform_get_ticks()) >= wolConnectDeadline {
			nox_client_setConnError_43AFA0(8)
		}
	case 4:
		sub_43AF90(3)
		sub_449E30(wolWindowString("TestCon", 1343))
		Set_qword_5d4594_814956(uint64(nox_platform_get_ticks() + 20000))
	case 5:
		sub_449E00(wolWindowString("Password", 1349))
		sub_449E30(wolWindowString("PasswordRequired", 1350))
		sub_449EA0(7)
		sub_44A360(0)
		sub_43AF90(6)
		sub_4A24C0(1)
	case 7:
		sub_44A360(1)
		sub_449E30(wolWindowString("Connected", 1364))
		sub_449EA0(0)
		nox_game_SetCliDrawFunc(C.nox_xxx_cliDrawConnectedLoop_43B360)
		sub_43AF90(1)
	case 8:
		ticks := uint64(nox_platform_get_ticks())
		sub_43AF90(9)
		*memmap.PtrUint64(0x5D4594, 814972) = ticks + 1000
	case 9:
		if uint64(nox_platform_get_ticks()) > memmap.Uint64(0x5D4594, 814972) {
			nox_client_joinGame_438A90()
		}
	case 10:
		password := (*wchar2_t)(unsafe.Pointer(uintptr(sub_449E60(4))))
		button := nox_xxx_wndGetChildByID_46B0C0(nil, 4001)
		nox_xxx_wnd_46ABB0_go(button, int32(bool2int(password != nil && *password != 0)))
	}
	return 1
}

//export sub_438BD0
func sub_438BD0() C.int {
	if sub_44A4A0() == 0 {
		nox_xxx_dialogMsgBoxCreate_449A10(wolWindows.root, nil, nil, 0, nil, nil)
	}
	code := wolConnectionError
	if code != 8 && code != 9 && code != 10 {
		sub_449E00(wolWindowString("ConnError", 1262))
		code = wolConnectionError
	}
	ids := [...]string{
		"ErrLowPing", "ErrHighPing", "ErrLowLevel", "ErrHighLevel",
		"ErrClosed", "ErrBanned", "ErrWrongPassword", "ErrIllegalClass",
		"ErrTimeOut", "ErrFindFailed", "ErrNeedRefresh", "ErrFull",
		"ErrDupSerial", "ErrWrongVer",
	}
	message := "ConnError"
	if uint(code) < uint(len(ids)) {
		message = ids[code]
	}
	sub_449E30(wolWindowString(message, 1265))
	Set_dword_5d4594_815044(0)
	sub_449EA0(1)
	sub_44A360(1)
	return C.int(sub_4A24C0(1))
}

//export sub_43A810
func sub_43A810() {
	nox_window_set_hidden(wolWindows.mapPanel, 0)
	nox_window_set_hidden(wolWindows.regionPanel, 1)
	region := C.int(Get_dword_587000_87412())
	nox_xxx_wndSetIcon_46AE60(wolWindows.mapPanel, nox_wol_map_icon(region))
	id, line := "JoinServer", int32(382)
	if Get_nox_game_createOrJoin_815048() == 1 {
		id, line = "CreateMsg", 377
	}
	wolSetHeading(wolWindows.heading, wolWindowString(id, line))
	sub_49FDB0(C.int(Get_dword_587000_87412()))
}

//export sub_43B630
func sub_43B630() *C.uint32_t {
	nox_xxx_wnd_46C6E0(wolWindows.serverInfo)
	message := wolWindowString("AttemptingConn", 1284)
	return (*C.uint32_t)(nox_xxx_dialogMsgBoxCreate_449A10(wolWindows.root, nil, message, 34, nil, nil))
}

func wolDisconnectedNotice(id string, line int32, pending *C.uint32_t) {
	if wolWindows.root == nil {
		*pending = 1
		return
	}
	message := wolWindowString(id, line)
	title := wolWindowString("Notification", line)
	nox_xxx_dialogMsgBoxCreate_449A10(nil, title, message, 33, nil, nil)
	sub_44A360(1)
	*pending = 0
}

//export sub_43B6E0
func sub_43B6E0() {
	wolDisconnectedNotice("Kicked", 3331, &C.dword_5d4594_815096)
}

//export sub_43B750
func sub_43B750() {
	wolDisconnectedNotice("Timeout", 3351, &C.dword_5d4594_815100)
}

//export nox_gui_wol_gameModeString_43BCB0
func nox_gui_wol_gameModeString_43BCB0(mode C.short) *C.wchar2_t {
	flags := noxflags.GameFlag(uint16(mode))
	for _, choice := range [...]struct {
		flags noxflags.GameFlag
		id    string
		line  int32
	}{
		{noxflags.GameModeQuest, "Quest", 3534},
		{noxflags.GameModeCTF, "CTF", 3541},
		{noxflags.GameModeElimination, "Highlander", 3545},
		{noxflags.GameModeKOTR, "KotR", 3550},
		{noxflags.GameModeFlagBall, "Flagball", 3555},
		{noxflags.GameModeChat, "Chat", 3559},
	} {
		if flags.Has(choice.flags) {
			return wolWindowString(choice.id, choice.line)
		}
	}
	return wolWindowString("Arena", 3563)
}
