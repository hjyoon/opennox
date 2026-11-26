package legacy

/*
#include "client__gui__guicon.h"
*/
import "C"

import (
	"strings"
	"unsafe"
)

func consoleLegacyString(key string, line int32) *wchar2_t {
	return nox_strman_loadString_40F1D0(internCStr(key), nil,
		internCStr(`C:\NoxPost\src\Client\System\parsecmd.c`), line)
}

//export nox_cmd_set_cycle
func nox_cmd_set_cycle(i, n C.int, tokens **C.wchar2_t) C.int {
	if n != 3 {
		return 0
	}
	switch strings.ToLower(GoWString(consoleLegacyToken(tokens, i, n))) {
	case "on":
		sub_4D0D90(1)
		consolePrintRed(consoleLegacyString("MapCycleOn", 2614))
	case "off":
		sub_4D0D90(0)
		consolePrintRed(consoleLegacyString("MapCycleOff", 2621))
	default:
		return 0
	}
	Sub_4AD840()
	return 1
}

func consoleSetEquipmentRule(token *wchar2_t, bit C.int, label string, lineOn, lineOff int32) C.int {
	var state string
	var line int32
	switch strings.ToLower(GoWString(token)) {
	case "on":
		sub_409E70(bit)
		state, line = "cmd_token:on", lineOn
	case "off":
		sub_409EC0(bit)
		state, line = "cmd_token:off", lineOff
	default:
		return 0
	}
	nox_server_gameSettingsUpdated_40A670()
	value := consoleLegacyString(state, line)
	format := consoleLegacyString(label, line+1)
	consolePrintRed(format, formatValue{text: unsafe.Pointer(value)})
	return 1
}

//export nox_cmd_set_weapons
func nox_cmd_set_weapons(i, n C.int, tokens **C.wchar2_t) C.int {
	if n != 3 {
		return 0
	}
	return consoleSetEquipmentRule(consoleLegacyToken(tokens, i, n), 1, "weapons", 2644, 2652)
}

//export nox_cmd_set_staffs
func nox_cmd_set_staffs(i, n C.int, tokens **C.wchar2_t) C.int {
	if n != 3 {
		return 0
	}
	return consoleSetEquipmentRule(consoleLegacyToken(tokens, i, n), 16, "staffs", 2674, 2682)
}

//export nox_cmd_set_players
func nox_cmd_set_players(i, n C.int, tokens **C.wchar2_t) C.int {
	if n != 3 {
		return 0
	}
	var limit C.int
	if token := consoleLegacyToken(tokens, 2, n); token != nil {
		limit = min(999, max(0, C.int(parseLegacyWideDecimal(token))))
		if nox_xxx_servGetPlrLimit_409FA0() == limit {
			return 1
		}
		nox_xxx_servSetPlrLimit_409F80(limit)
		sub_455800()
	}
	centerMessageWithNumber(consoleLegacyString("playersset", 3779), limit)
	return 1
}

//export nox_cmd_set_spellpts
func nox_cmd_set_spellpts(i, n C.int, tokens **C.wchar2_t) C.int {
	consolePrintRed(consoleLegacyString("notyetimplemented", 3830))
	return 1
}
