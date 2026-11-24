package legacy

/*
#include "client__network__deathmsg.h"
*/
import "C"

import (
	"slices"
	"unsafe"

	"github.com/opennox/libs/console"

	"github.com/opennox/opennox/v1/common/memmap"
)

func deathMessagePlayer(code C.uint16_t) *nox_playerInfo {
	if code == 0 {
		return nil
	}
	return nox_common_playerInfoGetByID_417040(int32(code))
}

//export sub_4952E0
func sub_4952E0(data *C.uint16_t) C.int {
	words := unsafe.Slice(data, 4)
	path := internCStr(`C:\NoxPost\src\client\Network\deathmsg.c`)
	attackerKey := internCStr("die.c:LocalizeAttacker")
	var attacker [128]wchar2_t
	if player := deathMessagePlayer(words[1]); player != nil {
		format := nox_strman_loadString_40F1D0(attackerKey, nil, path, 48)
		formatWideStrings(attacker[:], format, &player.name_final[0], nil)
		if player := deathMessagePlayer(words[2]); player != nil {
			n := slices.Index(attacker[:], 0)
			formatWideStrings(attacker[n:], internWStr(" + %s"), &player.name_final[0], nil)
		}
	} else {
		nasty := nox_strman_loadString_40F1D0(internCStr("die.c:AttackerNasty"), nil, path, 65)
		format := nox_strman_loadString_40F1D0(attackerKey, nil, path, 64)
		formatWideStrings(attacker[:], format, nasty, nil)
	}
	var victim [64]wchar2_t
	if player := deathMessagePlayer(words[3]); player != nil {
		format := nox_strman_loadString_40F1D0(internCStr("die.c:LocalizeVictim"), nil, path, 72)
		formatWideStrings(victim[:], format, &player.name_final[0], nil)
	}

	var message [512]wchar2_t
	format := (*wchar2_t)(memmap.PtrOff(0x587000, 161668))
	formatWideStrings(message[:], format, &victim[0], &attacker[0])
	return C.int(nox_gui_console_Print_450B90(C.uchar(console.ColorWhite), &message[0]))
}
