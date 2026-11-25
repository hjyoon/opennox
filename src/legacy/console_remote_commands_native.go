package legacy

/*
#include "client__system__parsecmd.h"
#include "client__gui__guicon.h"
#include "GAME3_2.h"

static void nox_console_print_remote_line(wchar2_t* text) {
	nox_gui_console_Printf_450C00(NOX_CONSOLE_RED, text);
}
static void nox_console_print_remote_value(wchar2_t* format, wchar2_t* value) {
	nox_gui_console_Printf_450C00(NOX_CONSOLE_RED, format, value);
}
static void nox_console_print_remote_command(wchar2_t* format, wchar2_t* sender, wchar2_t* text) {
	nox_gui_console_Printf_450C00(NOX_CONSOLE_RED, format, sender, text);
}
*/
import "C"

import (
	"strings"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//export nox_xxx_serverHandleClientConsole_443E90
func nox_xxx_serverHandleClientConsole_443E90(pl *C.nox_playerInfo, action C.char, text *C.wchar2_t) C.int {
	if pl == nil || pl.playerUnit == nil {
		return 0
	}
	if action != 0 && action != 4 && action != 5 && noxflags.HasGame(noxflags.GameFlag15|noxflags.GameFlag16) {
		return 1
	}
	switch action {
	case 0:
		consolePlayerObserve(pl, GoWString(text))
	case 1:
		if consoleRemoteAuthorized(pl) {
			consoleRunScript(pl, GoWString(text))
		}
	case 2:
		if consoleRemoteAuthorized(pl) {
			C.nox_console_print_remote_command(consoleLegacyString("RemoteSysop", 4345), &pl.name_final[0], text)
			nox_server_parseCmdText_443C80(text, 0)
		}
	case 3:
		consoleBroadcastText(0, text)
	case 4:
		consoleCameraFollow(pl, text)
	case 5:
		consoleBroadcastText(16, text)
		for peer := nox_common_playerInfoGetFirst_416EA0(); peer != nil; peer = nox_common_playerInfoGetNext_416EE0(peer) {
			if peer.playerUnit != nil {
				nox_xxx_aud_501960(902, peer.playerUnit, 0, 0)
			}
		}
	default:
		C.nox_console_print_remote_command(consoleLegacyString("invalidattempt", 4400), &pl.name_final[0], text)
	}
	return 1
}

func consoleBroadcastText(tag C.char, text *wchar2_t) {
	var words [256]uint16
	alloc.StrCopy16(words[:], GoWString(text))
	nox_broadcast_player_text_4D9FD0(tag, (*wchar2_t)(unsafe.Pointer(&words[0])))
}

func consoleRemoteAuthorized(pl *nox_playerInfo) bool {
	return sub_4D12A0(C.int(pl.playerInd)) != 0 ||
		uint32(pl.playerInd) == uint32(ClientPlayerNetCode()) || noxflags.HasGame(noxflags.GameModeCoop)
}

func consolePlayerObserve(pl *nox_playerInfo, text string) {
	if noxflags.HasGame(noxflags.GameFlag4|noxflags.GameModeQuest) || pl.field_3680&1 != 0 {
		return
	}
	forced := text == "\uf00d\uf0ad"
	if nox_xxx_playerGoObserver_4E6860(pl, int32(bool2int(forced)), 0) != 1 {
		return
	}
	label := consoleLegacyString("set", 4287)
	if nox_xxx_gamePlayIsAnyPlayers_40A8A0() != 0 {
		if !forced {
			nox_xxx_netNeedTimestampStatus_4174F0(pl, 256)
		}
		if noxflags.HasGame(noxflags.GameModeElimination) && sub_40A770() == 1 {
			sub_5095E0()
		}
	}
	if unit := pl.playerUnit; unit != nil {
		nox_xxx_netChangeTeamMb_419570((*C.nox_object_team_t)(unsafe.Pointer(&unit.field_12)), C.int(pl.netCode))
	}
	C.nox_console_print_remote_value(consoleLegacyString("observermode", 4308), label)
}

func consoleRunScript(pl *nox_playerInfo, text string) {
	fields := strings.FieldsFunc(text, func(r rune) bool { return r == ' ' })
	var (
		name  *wchar2_t
		index int32 = -1
	)
	if len(fields) >= 2 {
		ptr, free := CWString(fields[1])
		defer free()
		name = ptr
		key, freeKey := alloc.CString(fields[1])
		defer freeKey()
		index = nox_script_indexByEvent((*C.char)(unsafe.Pointer(key)))
	}
	if index == -1 {
		C.nox_console_print_remote_value(consoleLegacyString("InvalidFunction", 4329), name)
		return
	}
	C.nox_console_print_remote_value(consoleLegacyString("ExecutingFunction", 4332), name)
	unit := unsafe.Pointer(pl.playerUnit)
	nox_script_callByIndex_507310(index, unit, unit)
}

func consoleCameraFollow(pl *nox_playerInfo, text *wchar2_t) {
	if pl.field_3680&1 == 0 && !noxflags.HasEngine(noxflags.EngineReplayRead) {
		if noxflags.HasGame(noxflags.GameHost) {
			C.nox_console_print_remote_line(consoleLegacyString("notinobserver", 4392))
		}
		return
	}
	if text == nil || *text == 0 {
		nox_xxx_playerCameraUnlock_4E6040(pl.playerUnit)
		return
	}
	for peer := nox_common_playerInfoGetFirst_416EA0(); peer != nil; peer = nox_common_playerInfoGetNext_416EE0(peer) {
		if C._nox_wcsicmp(text, &peer.name_final[0]) == 0 {
			nox_xxx_playerCameraFollow_4E6060(pl.playerUnit, peer.playerUnit)
		}
	}
}
