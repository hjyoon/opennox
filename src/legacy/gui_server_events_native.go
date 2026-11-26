package legacy

/*
#include <stdlib.h>
#include "cgo_const.h"
#include "game2_server_bridge.h"
#include "client__gui__servopts__playrlst.h"
#include "common__magic__speltree.h"
#include "common__system__team.h"

extern nox_window* dword_5d4594_1046500;
extern nox_window* dword_5d4594_1046524;
extern nox_window* dword_5d4594_1046528;
extern nox_window* dword_5d4594_1046532;
extern nox_window* dword_5d4594_1046536;
extern nox_window* dword_5d4594_1046540;
*/
import "C"

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func serverOptionsChild(id int32) *nox_window {
	return nox_xxx_wndGetChildByID_46B0C0(C.dword_5d4594_1046492, id)
}

func serverOptionsListChanged(window *nox_window, row uintptr) {
	switch nox_xxx_wndGetID_46B0A0(window) {
	case 10114:
		if int(g2ServerEvent(window, 16404, 0, 0)) < 0 {
			return
		}
		name := serverOptionMapNameFromRow(listBoxRowText(window, row))
		record := sub_4165B0()
		settings := (*server.Settings2)(unsafe.Pointer(record))
		alloc.StrCopy(settings.Field0[:9], name)
		readMapRuleDefaults57A1E0(settings, internCStr("user.rul"), nil, ruleSourceMap|ruleSourceUser|ruleSourceInternet, settings.Field52)
		sub_459880(record)
		sub_459D50(1)
	case 10120:
		record := sub_4165B0()
		modeButton := serverOptionsChild(10119)
		selected := int(g2ServerEvent(window, 16404, 0, 0))
		data := g2ServerList(window)
		if data == nil || selected < 0 || selected >= int(data.Field_11_0) {
			return
		}
		settings := (*server.Settings2)(unsafe.Pointer(record))
		title := listBoxRowText(window, row)
		g2ServerText(modeButton, 16385, unsafe.Pointer(title), ^uintptr(0))
		mask := uint16(noxflags.GameModeMask)
		oldMode := settings.Field52 & mask
		settings.Field52 = settings.Field52&^mask | uint16(sub_459650(title))
		if settings.Field52&mask != oldMode {
			nox_client_guiserv_updateMapList_458230(C.int(settings.Field52), (*C.char)(g2ServerMem(1046556)), true)
		}
		g2ServerHidden(window, true)
		C.nox_xxx_wndClearCaptureMain_46ADE0(window)
		settings.Field54 = uint16(nox_xxx_servGamedataGet_40A020(C.short(settings.Field52)))
		settings.Field56 = uint8(sub_40A180(C.short(settings.Field52)))
		sub_457460((*C.nox_cgo_const_char)(unsafe.Pointer(record)))
		sub_459D50(1)
		if noxflags.GameFlag(settings.Field52).Has(noxflags.GameModeQuest) {
			C.nox_xxx_spellEnableAll_424BD0()
			sub_4537F0()
			C.sub_459700()
		}
		sub_457F30(g2ServerBool(noxflags.GameFlag(settings.Field52).Has(noxflags.GameModeQuest)))
	}
}

func serverOptionsApply() {
	record := C.nox_xxx_cliGamedataGet_416590(1)
	nox_game2_server_apply_459AA0(unsafe.Pointer(record))
	sub_4165F0(1, 0)
	count := int(nox_xxx_getTeamCounter_417DD0())
	root := C.dword_5d4594_1046492
	if noxflags.HasGame(noxflags.GameModeChat) {
		mode := (*server.Settings2)(unsafe.Pointer(record)).Field52
		lastLine := 0
		if mode&uint16(noxflags.GameModeCTF|noxflags.GameModeFlagBall) != 0 {
			if int16(mode) >= 0 && count < 2 {
				message := g2ServerLoad("NeedTeams", serverOptionsStringsFile, 1363)
				title := serverOptionsText(1046560)
				nox_xxx_dialogMsgBoxCreate_449A10(root, title, message, 56, unsafe.Pointer(C.sub_459150), nil)
				C.sub_44A360(1)
				return
			}
			if count > 2 {
				lastLine = 1371
			}
		} else if noxflags.HasGame(noxflags.GameModeKOTR) && bool(nox_xxx_CheckGameplayFlags_417DA0(4)) && count > 2 {
			lastLine = 1384
		}
		if lastLine != 0 {
			message := g2ServerLoad("TooManyTeams", serverOptionsStringsFile, lastLine)
			title := g2ServerLoad("Notice", serverOptionsStringsFile, lastLine-1)
			nox_xxx_dialogMsgBoxCreate_449A10(root, title, message, 33, nil, nil)
			C.sub_44A360(1)
			return
		}
	}
	sub_459150()
}

func serverOptionsSwitchTab(window *nox_window, showTeams bool) {
	root := C.dword_5d4594_1046492
	parent := window.parent
	sub_46B120(window, nil)
	sub_46B120(window, parent)
	if showTeams {
		if noxflags.HasGame(noxflags.GameHost) {
			g2ServerHidden(serverOptionsChild(10196), false)
			sub_46ACE0(root, 10161, 10163, 0)
			sub_46AD20(root, 10161, 10163, 1)
			C.sub_4593B0(0)
			serverOptionsChild(10163).draw_data.field_0 |= 4
			serverOptionsChild(10161).draw_data.field_0 &^= 4
			serverOptionsChild(10162).draw_data.field_0 &^= 4
		} else {
			g2ServerHidden(serverOptionsChild(10196), true)
			C.dword_5d4594_1046532 = nox_xxx_guiServerPlayersLoad_456270(root)
		}
		sub_46ACE0(root, 10141, 10141, 1)
		g2ServerHidden(C.dword_5d4594_1046524, true)
		return
	}
	if C.dword_5d4594_1046532 != nil {
		C.sub_456D60(1)
		C.dword_5d4594_1046532 = nil
	}
	if C.dword_5d4594_1046528 != nil {
		C.sub_4557D0(1)
		C.dword_5d4594_1046528 = nil
	}
	if C.dword_5d4594_1046536 != nil {
		C.sub_4AD820()
		C.dword_5d4594_1046540 = nil
	}
	sub_46ACE0(root, 10161, 10163, 1)
	if noxflags.HasGame(noxflags.GameHost) {
		sub_46ACE0(root, 10141, 10141, 0)
	}
	g2ServerHidden(serverOptionsChild(10196), true)
	g2ServerHidden(C.dword_5d4594_1046524, false)
}

func serverOptionsButton(window *nox_window) {
	id := nox_xxx_wndGetID_46B0A0(window)
	nox_xxx_clientPlaySoundSpecial_452D80(766, 100)
	switch id {
	case 10119:
		modes := serverOptionsChild(10120)
		g2ServerHidden(modes, false)
		C.nox_xxx_wndSetCaptureMain_46ADC0(modes)
		C.nox_xxx_wndShowModalMB_46A8C0(modes)
		nox_xxx_windowServerOptionsFillGametypeList_4596A0()
	case 10122:
		enable := window.draw_data.field_0&4 != 0
		if !g2ServerGame(uint32(serverOptionsIndividualLadder | serverOptionsClanLadder)) {
			g2ServerEnabled(C.dword_5d4594_1046500, enable)
			g2ServerEnabled(C.dword_5d4594_1046504, enable)
		}
		sub_459D50(1)
	case 10141:
		C.sub_459700()
	case 10145:
		serverOptionsApply()
	case 10146:
		settings := (*server.Settings2)(unsafe.Pointer(sub_4165B0()))
		if noxflags.HasGame(noxflags.GameModeChat) {
			if int16(settings.Field52) < 0 {
				C.nox_server_teamsZzz_419030(1)
			}
			settings.Field52 &= 0x3fff
		}
		nox_xxx_guiServerOptionsHide_4597E0(0)
	case 10149:
		nox_xxx_guiServerOptionsHide_4597E0(0)
	case 10152:
		C.nox_xxx_loadAdvancedWnd_4BDC10((*C.int)(unsafe.Pointer(C.nox_xxx_cliGamedataGet_416590(1))))
	case 10159:
		serverOptionsSwitchTab(window, true)
	case 10160:
		serverOptionsSwitchTab(window, false)
	case 10161:
		C.sub_4593B0(1)
	case 10162:
		C.sub_4593B0(2)
	case 10163:
		C.sub_4593B0(0)
	case 10330:
		panel := C.dword_5d4594_1046508
		if bool(nox_xxx_CheckGameplayFlags_417DA0(4)) {
			nox_xxx_toggleAllTeamFlags_418690(0)
			C.nox_server_teamsZzz_419030(1)
			sub_46AD20(panel, 10331, 10333, 0)
		} else {
			nox_xxx_wndGuiTeamCreate_4185B0()
			if bool(nox_xxx_CheckGameplayFlags_417DA0(2)) {
				C.sub_4181F0(0)
			} else {
				nox_xxx_toggleAllTeamFlags_418690(1)
			}
			sub_46AD20(panel, 10331, 10333, 1)
		}
	case 10331:
		if bool(nox_xxx_CheckGameplayFlags_417DA0(2)) {
			C.nox_xxx_UnsetGameplayFlags_417D70(2)
			nox_xxx_toggleAllTeamFlags_418690(1)
		} else {
			nox_xxx_toggleAllTeamFlags_418690(0)
			C.sub_418390()
		}
	case 10332:
		C.sub_4181F0(1)
	case 10333:
		if bool(nox_xxx_CheckGameplayFlags_417DA0(1)) {
			C.nox_xxx_UnsetGameplayFlags_417D70(1)
		} else {
			C.nox_xxx_SetGameplayFlag_417D50(1)
		}
	}
}

func serverOptionsFieldChanged(eventArg, id uintptr) C.int {
	window := serverOptionsChild(int32(id))
	settings := (*server.Settings2)(unsafe.Pointer(sub_4165B0()))
	if window == nil {
		return 0
	}
	if uint16(eventArg) == 1 {
		if id == 10101 {
			window.draw_data.text_color = C.uint32_t(Get_nox_color_white_2523948())
		}
		return 1
	}
	text := (*C.wchar2_t)(unsafe.Pointer(g2ServerEvent(window, 16413, 0, 0)))
	if text == nil {
		return 1
	}
	var narrow [200]byte
	copyLegacyCString(narrow[:], (*uint16)(unsafe.Pointer(text)))
	if narrow[0] == 0 {
		return 1
	}
	switch id {
	case 10101:
		window.draw_data.text_color = C.uint32_t(nox_color_rgb_4344A0(230, 165, 65))
		alloc.StrCopy(settings.Field0[9:], string(narrow[:StrLenBytes(narrow[:])]))
		nox_xxx_gameSetServername_40A440((*C.char)(unsafe.Pointer(&settings.Field0[9])))
	case 10134:
		settings.Field54 = uint16(max(parseLegacyDecimal32((*C.char)(unsafe.Pointer(&narrow[0]))), 0))
		sub_459D50(1)
	case 10135:
		value := max(parseLegacyDecimal32((*C.char)(unsafe.Pointer(&narrow[0]))), 0)
		if value > 255 {
			value = 255
			g2ServerIntText(window, 255, ^uintptr(0))
		}
		settings.Field56 = uint8(value)
		sub_459D50(1)
	}
	return 1
}

//export nox_xxx_guiServerOptionsProcPre_4585D0
func nox_xxx_guiServerOptionsProcPre_4585D0(_ *C.nox_window, event C.uint, first, second C.uintptr_t) C.int {
	switch event {
	case 23:
		return 0
	case 16387:
		return serverOptionsFieldChanged(uintptr(first), uintptr(second))
	case 16391:
		serverOptionsButton((*nox_window)(unsafe.Pointer(uintptr(first))))
	case 16400:
		serverOptionsListChanged((*nox_window)(unsafe.Pointer(uintptr(first))), uintptr(second))
	}
	return 1
}
