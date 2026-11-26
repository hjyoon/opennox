package legacy

/*
#include "GAME1.h"
#include "GAME2.h"
#include "client__gui__servopts__guiserv.h"
#include "common__system__settings.h"

*/
import "C"

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

func consoleLegacyToken(tokens **wchar2_t, i, n C.int) *wchar2_t {
	return unsafe.Slice(tokens, int(n))[i]
}

//export nox_cmd_offonly1
func nox_cmd_offonly1(i, n C.int, tokens **C.wchar2_t) C.int {
	if n != 2 {
		return 0
	}
	Nox_xxx_wndGuiTeamCreate_4185B0()
	return 1
}

//export nox_cmd_set_net_debug
func nox_cmd_set_net_debug(i, n C.int, tokens **C.wchar2_t) C.int {
	if n != 2 {
		return 0
	}
	noxflags.SetEngine(noxflags.EngineNetDebug)
	return 1
}

//export nox_cmd_unset_net_debug
func nox_cmd_unset_net_debug(i, n C.int, tokens **C.wchar2_t) C.int {
	if n != 2 {
		return 0
	}
	noxflags.UnsetEngine(noxflags.EngineNetDebug)
	return 1
}

//export nox_cmd_show_info
func nox_cmd_show_info(i, n C.int, tokens **C.wchar2_t) C.int {
	nox_xxx_clientPlaySoundSpecial_452D80(921, 100)
	sub_435F60()
	return 1
}

//export nox_cmd_show_mem
func nox_cmd_show_mem(i, n C.int, tokens **C.wchar2_t) C.int {
	nox_server_currentMapGetFilename_409B30()
	nox_xxx_gameLoopMemDump_413E30()
	return 1
}

//export nox_cmd_show_rank
func nox_cmd_show_rank(i, n C.int, tokens **C.wchar2_t) C.int {
	if noxflags.HasGame(noxflags.GameOnline) {
		sub_4703F0()
	}
	return 1
}

//export nox_cmd_show_motd
func nox_cmd_show_motd(i, n C.int, tokens **C.wchar2_t) C.int {
	if n != 2 {
		return 0
	}
	nox_xxx_motd_4467F0()
	return 1
}

//export nox_cmd_show_seq
func nox_cmd_show_seq(i, n C.int, tokens **C.wchar2_t) C.int {
	if n != 2 {
		return 0
	}
	sub_48D7B0()
	return 1
}

//export nox_cmd_cheat_ability
func nox_cmd_cheat_ability(i, n C.int, tokens **C.wchar2_t) C.int {
	if !noxflags.HasGame(noxflags.GameOnline) {
		players := &GetServer().S().Players
		for pl := players.First(); pl != nil; pl = players.Next(pl) {
			if pl.PlayerUnit != nil {
				nox_xxx_playerCancelAbils_4FC180(asObjectC(pl.PlayerUnit))
			}
		}
	}
	return 1
}

//export nox_cmd_cheat_level
func nox_cmd_cheat_level(i, n C.int, tokens **C.wchar2_t) C.int {
	if noxflags.HasGame(noxflags.GameOnline) {
		return 1
	}
	if n < 3 {
		return 0
	}
	players := &GetServer().S().Players
	for pl := players.First(); pl != nil; pl = players.Next(pl) {
		if pl.PlayerUnit != nil {
			level := C.uint8_t(parseLegacyWideDecimal(consoleLegacyToken(tokens, 2, n)))
			sub_4EF410(asObjectC(pl.PlayerUnit), level)
		}
	}
	return 1
}

//export nox_cmd_window
func nox_cmd_window(i, n C.int, tokens **C.wchar2_t) C.int {
	if n > 1 {
		token := consoleLegacyToken(tokens, i, n)
		value := parseLegacyWideDecimal(token)
		if *token == '+' || *token == '-' {
			nox_draw_setCutSize_476700(0, int32(value))
		} else {
			nox_draw_setCutSize_476700(C.int(value), 0)
		}
	}
	return 1
}

func consoleSetConnectionQuality(rate C.int, internet bool) C.int {
	legacyWords.dword_5d4594_2650652 = C.uint32_t(bool2int(internet))
	nox_xxx_rateUpdate_40A6D0(sub_40A710(rate))
	Set_nox_server_connectionType_3596(int(rate))
	Sub_4AD840()
	return 1
}

//export nox_cmd_set_qual_modem
func nox_cmd_set_qual_modem(i, n C.int, tokens **C.wchar2_t) C.int {
	return consoleSetConnectionQuality(4, true)
}

//export nox_cmd_set_qual_isdn
func nox_cmd_set_qual_isdn(i, n C.int, tokens **C.wchar2_t) C.int {
	return consoleSetConnectionQuality(3, true)
}

//export nox_cmd_set_qual_cable
func nox_cmd_set_qual_cable(i, n C.int, tokens **C.wchar2_t) C.int {
	return consoleSetConnectionQuality(2, true)
}

//export nox_cmd_set_qual_t1
func nox_cmd_set_qual_t1(i, n C.int, tokens **C.wchar2_t) C.int {
	return consoleSetConnectionQuality(1, true)
}

//export nox_cmd_set_qual_lan
func nox_cmd_set_qual_lan(i, n C.int, tokens **C.wchar2_t) C.int {
	return consoleSetConnectionQuality(1, false)
}

func consoleSettingsMode(settings *C.char) C.short {
	return C.short((*server.Settings2)(unsafe.Pointer(settings)).Field52)
}

//export nox_cmd_set_time
func nox_cmd_set_time(i, n C.int, tokens **C.wchar2_t) C.int {
	if n != 3 {
		return 0
	}
	settings := sub_4165B0()
	if token := consoleLegacyToken(tokens, 2, n); token != nil {
		time := C.uchar(parseLegacyWideDecimal(token))
		sub_40A040_settings(consoleSettingsMode(settings), time)
	}
	return 1
}

//export nox_cmd_set_lessons
func nox_cmd_set_lessons(i, n C.int, tokens **C.wchar2_t) C.int {
	if n != 3 {
		return 0
	}
	token := consoleLegacyToken(tokens, 2, n)
	settings := sub_4165B0()
	if token != nil {
		lessons := C.ushort(parseLegacyWideDecimal(token))
		sub_409FB0_settings(consoleSettingsMode(settings), lessons)
	}
	return 1
}

//export nox_cmd_menu_options
func nox_cmd_menu_options(i, n C.int, tokens **C.wchar2_t) C.int {
	if !noxflags.HasGame(noxflags.GameFlag4) && noxflags.HasGame(noxflags.GameOnline) {
		C.nox_xxx_guiServerOptsLoad_457500()
	}
	return 1
}

//export nox_cmd_menu_vidopt
func nox_cmd_menu_vidopt(i, n C.int, tokens **C.wchar2_t) C.int {
	sub_4ADA40()
	return 1
}

//export nox_cmd_reenter
func nox_cmd_reenter(i, n C.int, tokens **C.wchar2_t) C.int {
	if !noxflags.HasGame(noxflags.GameOnline) {
		sub_40AA60(1)
	}
	return 1
}
