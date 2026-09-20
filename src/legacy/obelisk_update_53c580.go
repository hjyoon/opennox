package legacy

/*
#include <stdint.h>

#include "GAME5_2.h"

void nullsub_36(void);

static inline void nox_obelisk_protect_mana_53C580(int token, int16_t delta) {
	(void)nox_xxx_protectMana_56F9E0(token, (short)delta);
}

static inline void nox_obelisk_protect_max_mana_53C580(int token, uint16_t value) {
	(void)nox_xxx_protectPlayerHPMana_56F870(token, (unsigned short)value);
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func obeliskUpdateCall53C580(obelisk *server.Object) {
	s := GetServer().S()
	s.ObeliskUpdate53C580(obelisk, server.ObeliskUpdateRuntime53C580{
		ReplenishmentEffect: C.nullsub_36,
		ProtectMana: func(token uint32, delta int16) {
			C.nox_obelisk_protect_mana_53C580(C.int(int32(token)), C.int16_t(delta))
		},
		ProtectPlayerHPMana: func(token uint32, value uint16) {
			C.nox_obelisk_protect_max_mana_53C580(C.int(int32(token)), C.uint16_t(value))
		},
		ReportCharges: Nox_xxx_netReportCharges_4D82B0,
	})
}
