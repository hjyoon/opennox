package legacy

/*
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_1.h"
#include "client__gui__guiquit.h"
#include "client__gui__window.h"

*/
import "C"

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

const escapeMenuStringsFile = "C:\\NoxPost\\src\\client\\Gui\\guiquit.c"

func escapeMenuText(key string, line int32) *C.wchar2_t {
	return nox_strman_loadString_40F1D0(internCStr(key), nil, internCStr(escapeMenuStringsFile), line)
}

//export nox_xxx_quitMenuCanAutoSave_445830
func nox_xxx_quitMenuCanAutoSave_445830() C.int {
	local := (*nox_drawable)(*memmap.PtrPtr(0x852978, 8))
	return C.int(bool2int(local != nil && local.flags30&0x8000 == 0))
}

//export nox_xxx_menuGameOnButton_445840
func nox_xxx_menuGameOnButton_445840(root *C.nox_window, event C.int, control *C.nox_window, unusedArg C.uintptr_t) C.int {
	if event != 16391 {
		return 0
	}
	id := nox_xxx_wndGetID_46B0A0(control)
	nox_xxx_clientPlaySoundSpecial_452D80(766, 100)
	switch id {
	case 9001:
		sub_445C40()
		sub_413A00(1)
		if !noxflags.HasGame(noxflags.GameModeCoop) || nox_xxx_playerAnimCheck_4372B0() != 0 {
			sub_445B40()
		} else {
			message := escapeMenuText("GUIQuit.c:ReallyLoadMessage", 199)
			title := escapeMenuText("SelChar.c:LoadLabel", 198)
			nox_xxx_dialogMsgBoxCreate_449A10(nil, title, message, 56,
				unsafe.Pointer(C.sub_445B40), unsafe.Pointer(C.sub_445BA0))
		}
	case 9002:
		if nox_xxx_quitMenuCanAutoSave_445830() != 0 {
			sub_445C40()
			if noxflags.HasGame(noxflags.GameModeCoop) {
				nox_setSaveFileName_4DB130(internCStr("AUTOSAVE"))
				sub_4DB170(1, nil, 0)
			}
		} else {
			g2ServerEnabled(nox_xxx_wndGetChildByID_46B0C0(root, id), false)
		}
	case 9003:
		sub_445C40()
		if noxflags.HasGame(noxflags.GameModeCoop) {
			nox_savegame_sub_46D580()
		} else {
			nox_xxx_netSavePlayer_41CE00()
		}
		if sub_43C6E0() == 0 {
			sub_43CF70()
		}
	case 9004:
		menu := quitMenuWindow
		nox_xxx_wndClearCaptureMain_46ADE0(menu)
		message := escapeMenuText("GUIQuit.c:ReallyQuitMessage", 185)
		title := escapeMenuText("GUIQuit.c:ReallyQuitTitle", 184)
		nox_xxx_dialogMsgBoxCreate_449A10(menu, title, message, 56,
			unsafe.Pointer(C.nox_xxx_quitDialogYes_445B20), unsafe.Pointer(C.nox_xxx_quitDialogNo_445B30))
	case 9005:
		sub_445C40()
		sub_4ADA40()
	case 9006:
		sub_445C40()
	case 9007:
		if noxflags.HasGame(noxflags.GameHost) {
			player := nox_common_playerInfoGetByID_417040(int32(ClientPlayerNetCode()))
			nox_xxx_serverHandleClientConsole_443E90(player, 0, nil)
		} else {
			nox_xxx_netServerCmd_440950_empty()
		}
		sub_445C40()
	case 9008:
		sub_445C40()
		nox_xxx_guiServerOptsLoad_457500()
	case 9009:
		sub_445C40()
		if noxflags.HasGame(noxflags.GameModeQuest) {
			sub_48CB10(4)
		} else {
			sub_48CB10(0)
		}
	}
	if control != nil {
		control.draw_data.field_0 &^= 2
	}
	return 0
}

func showEscapeSolo(menu *nox_window) {
	saveLabel := escapeMenuText("SoloSaveLabel", 396)
	C.sub_46AEE0(nox_xxx_wndGetChildByID_46B0C0(menu, 9003), saveLabel)
	for _, id := range [...]int32{9001, 9002} {
		nox_window_set_hidden(nox_xxx_wndGetChildByID_46B0C0(menu, id), 0)
	}
	for _, id := range [...]int32{9007, 9008, 9009} {
		nox_window_set_hidden(nox_xxx_wndGetChildByID_46B0C0(menu, id), 1)
	}
	if team := nox_xxx_wndGetChildByID_46B0C0(menu, 9009); team != nil {
		quit := nox_xxx_wndGetChildByID_46B0C0(menu, 9004)
		nox_window_setPos_46A9B0(quit, int32(team.off_x), int32(team.off_y))
	}
	sub_413A00(1)
	sub_46AB20(menu, 220, 285)
}

func showEscapeMultiplayer(menu *nox_window) {
	saveLabel := escapeMenuText("MultiplayerSaveLabel", 427)
	C.sub_46AEE0(nox_xxx_wndGetChildByID_46B0C0(menu, 9003), saveLabel)
	for _, id := range [...]int32{9001, 9002} {
		nox_window_set_hidden(nox_xxx_wndGetChildByID_46B0C0(menu, id), 1)
	}
	vote := nox_xxx_wndGetChildByID_46B0C0(menu, 9007)
	nox_window_set_hidden(vote, 0)
	g2ServerEnabled(vote, true)
	nox_window_set_hidden(nox_xxx_wndGetChildByID_46B0C0(menu, 9008), 0)

	if team := nox_xxx_wndGetChildByID_46B0C0(menu, 9009); team != nil {
		nox_window_set_hidden(team, 0)
		C.sub_46AEE0(team, (*C.wchar2_t)(memmap.PtrOff(0x5D4594, 825772)))
		g2ServerEnabled(team, !noxflags.HasGame(noxflags.GameFlag15|noxflags.GameFlag16) && nox_xxx_getTeamCounter_417DD0() != 0)
		quit := nox_xxx_wndGetChildByID_46B0C0(menu, 9004)
		nox_window_setPos_46A9B0(quit, int32(team.off_x), int32(team.off_y)+45)
	}

	sub_46AB20(menu, 220, 330)
	if noxflags.HasGame(noxflags.GameModeQuest) {
		g2ServerEnabled(vote, false)
		g2ServerEnabled(nox_xxx_wndGetChildByID_46B0C0(menu, 9003), false)
	}
	if noxflags.HasEngine(noxflags.EngineNoRendering) {
		g2ServerEnabled(vote, false)
		g2ServerEnabled(nox_xxx_wndGetChildByID_46B0C0(menu, 9005), false)
		g2ServerEnabled(nox_xxx_wndGetChildByID_46B0C0(menu, 9003), false)
	}
}

//export sub_445C40
func sub_445C40() {
	menu := quitMenuWindow
	if menu == nil {
		return
	}
	if menu.flags&0x10 == 0 {
		nox_xxx_windowFocus_46B500(nil)
		nox_xxx_wndClearCaptureMain_46ADE0(menu)
		nox_window_set_hidden(menu, 1)
		menu.flags &^= 8
		sub_413A00(0)
		return
	}
	if local := (*nox_drawable)(*memmap.PtrPtr(0x852978, 8)); local != nil && noxflags.HasGame(noxflags.GameModeCoop) {
		switch local.field_69 {
		case 1, 2, 51:
			return
		}
	}
	if sub_45D9B0() == 1 || nox_xxx_checkGameFlagPause_413A50() == 1 {
		return
	}
	nox_xxx_clientPlaySoundSpecial_452D80(921, 100)
	nox_xxx_wndShowModalMB_46A8C0(menu)
	menu.flags |= 8
	nox_xxx_wndSetCaptureMain_46ADC0(menu)
	if noxflags.HasGame(noxflags.GameModeCoop) {
		showEscapeSolo(menu)
	} else {
		showEscapeMultiplayer(menu)
	}
}
