package legacy

/*
#include "buff_apply_4ff380.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func buffApplyRuntime4FF380() server.BuffApplyRuntime4FF380 {
	return server.BuffApplyRuntime4FF380{
		BuffOff: func(unit *server.Object, buff server.EnchantID) int32 {
			Nox_xxx_spellBuffOff_4FF5B0(unit, buff)
			return 0
		},
		ResetPlayerProtection: func(player *server.Player, flags uint32) {
			Nox_xxx_playerResetProtectionCRC_56F7D0(player.ProtUnitBuffs, int(flags))
		},
	}
}

func buffApplyLegacy4FF380(unit *server.Object, buff int32, duration int16, power int8) {
	GetServer().S().BuffApply4FF380(unit, buff, duration, power, buffApplyRuntime4FF380())
}

var buffApplyExportImpl4FF380 = buffApplyLegacy4FF380

func buffApplyExportCall4FF380(unit *server.Object, buff int32, duration int16, power int8) {
	C.nox_xxx_buffApplyTo_4FF380(
		asObjectC(unit),
		C.int32_t(buff),
		C.int16_t(duration),
		C.int8_t(power),
	)
}

//export nox_xxx_buffApplyTo_4FF380
func nox_xxx_buffApplyTo_4FF380(
	unit *C.nox_object_t,
	buff C.int32_t,
	duration C.int16_t,
	power C.int8_t,
) {
	buffApplyExportImpl4FF380(
		asObjectS((*nox_object_t)(unit)),
		int32(buff),
		int16(duration),
		int8(power),
	)
}
