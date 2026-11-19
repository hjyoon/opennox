package legacy

/*
#include "spell_buff_off_4ff5b0.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func spellBuffOffRuntime4FF5B0() server.SpellBuffOffRuntime4FF5B0 {
	return server.SpellBuffOffRuntime4FF5B0{
		ResetPlayerProtection: func(player *server.Player, flags uint32) {
			Nox_xxx_playerResetProtectionCRC_56F7D0(player.ProtUnitBuffs, int(flags))
		},
	}
}

func spellBuffOffLegacy4FF5B0(unit *server.Object, buff int32) int32 {
	return GetServer().S().SpellBuffOff4FF5B0(unit, buff, spellBuffOffRuntime4FF5B0())
}

var spellBuffOffExportImpl4FF5B0 = spellBuffOffLegacy4FF5B0

func spellBuffOffExportCall4FF5B0(unit *server.Object, buff int32) int32 {
	return int32(C.nox_xxx_spellBuffOff_4FF5B0(
		asObjectC(unit),
		C.int32_t(buff),
	))
}

//export nox_xxx_spellBuffOff_4FF5B0
func nox_xxx_spellBuffOff_4FF5B0(unit *C.nox_object_t, buff C.int32_t) C.int32_t {
	return C.int32_t(spellBuffOffExportImpl4FF5B0(
		asObjectS((*nox_object_t)(unit)),
		int32(buff),
	))
}

//export sub_5341D0
func sub_5341D0(unit *C.nox_object_t) C.int {
	obj := asObjectS((*nox_object_t)(unit))
	spellBuffOffExportImpl4FF5B0(obj, int32(server.ENCHANT_INVISIBLE))
	return C.int(spellBuffOffExportImpl4FF5B0(obj, int32(server.ENCHANT_INVULNERABLE)))
}
