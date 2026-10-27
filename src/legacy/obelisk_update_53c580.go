package legacy

/*
#include <stdint.h>

#include "GAME4_3.h"
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

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func obeliskUpdateRuntime53C580() server.ObeliskUpdateRuntime53C580 {
	return server.ObeliskUpdateRuntime53C580{
		ReplenishmentEffect: C.nullsub_36,
		ProtectMana: func(token uint32, delta int16) {
			C.nox_obelisk_protect_mana_53C580(C.int(int32(token)), C.int16_t(delta))
		},
		ProtectPlayerHPMana: func(token uint32, value uint16) {
			C.nox_obelisk_protect_max_mana_53C580(C.int(int32(token)), C.uint16_t(value))
		},
		ReportCharges: Nox_xxx_netReportCharges_4D82B0,
	}
}

func obeliskUpdateCall53C580(obelisk *server.Object) {
	GetServer().S().ObeliskUpdate53C580(obelisk, obeliskUpdateRuntime53C580())
}

//export nox_xxx_rechargeItem_53C520_go
func nox_xxx_rechargeItem_53C520_go(item *nox_object_t, amount C.int) C.int {
	if server.ObeliskRechargeItem53C520(asObjectS(item), int32(amount)) {
		return 1
	}
	return 0
}

//export nox_xxx_updateObelisk_53C580_go
func nox_xxx_updateObelisk_53C580_go(obelisk *nox_object_t) C.intptr_t {
	return C.intptr_t(GetServer().S().ObeliskUpdate53C580(asObjectS(obelisk), obeliskUpdateRuntime53C580()))
}

//export nox_xxx_getRechargeRate_53C940_go
func nox_xxx_getRechargeRate_53C940_go(item *nox_object_t) C.int {
	return C.int(GetServer().S().ObeliskRechargeRate53C940(asObjectS(item), obeliskUpdateRuntime53C580()))
}

func obeliskRechargeItemC53C520(item *server.Object, amount int32) int32 {
	return int32(C.nox_xxx_rechargeItem_53C520(asObjectC(item), C.int(amount)))
}

func obeliskUpdateC53C580(obelisk *server.Object) int {
	return int(C.nox_xxx_updateObelisk_53C580(asObjectC(obelisk)))
}

func obeliskRechargeRateC53C940(item *server.Object) int32 {
	return int32(C.nox_xxx_getRechargeRate_53C940(asObjectC(item)))
}

func obeliskReplenishmentEffectC53C940() unsafe.Pointer {
	return C.nullsub_36
}
