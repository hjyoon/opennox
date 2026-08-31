package legacy

/*
#include <stdint.h>

#include "GAME4_3.h"
#include "server__ability__ability.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func abilityRewardServLegacy4FB9C0(
	unit *server.Object,
	ability, rewardArg int32,
) int32 {
	return GetServer().AbilityRewardServ4FB9C0(unit, ability, rewardArg)
}

func useAbilityRewardLegacy53FAE0(owner, item *server.Object) int32 {
	return GetServer().UseAbilityReward53FAE0(owner, item)
}

func abilityRewardExportCall4FB9C0(
	unit *server.Object,
	ability, rewardArg int32,
) int32 {
	return int32(C.nox_xxx_abilityRewardServ_4FB9C0_ability(
		asObjectC(unit),
		C.int32_t(ability),
		C.int32_t(rewardArg),
	))
}

func useAbilityRewardExportCall53FAE0(owner, item *server.Object) int32 {
	return int32(C.nox_xxx_useAbilityReward_53FAE0(
		asObjectC(owner),
		asObjectC(item),
	))
}

// Nox_xxx_abilityRewardServ_4FB9C0_ability gives Go-owned callers the
// restored native-pointer service path.
func Nox_xxx_abilityRewardServ_4FB9C0_ability(
	unit *server.Object,
	ability, rewardArg int32,
) int32 {
	return abilityRewardServLegacy4FB9C0(unit, ability, rewardArg)
}

//export nox_xxx_abilityRewardServ_4FB9C0_ability
func nox_xxx_abilityRewardServ_4FB9C0_ability(
	unit *C.nox_object_t,
	ability, rewardArg C.int32_t,
) C.int32_t {
	return C.int32_t(abilityRewardServLegacy4FB9C0(
		asObjectS((*nox_object_t)(unit)),
		int32(ability),
		int32(rewardArg),
	))
}

//export nox_xxx_useAbilityReward_53FAE0
func nox_xxx_useAbilityReward_53FAE0(
	owner, item *C.nox_object_t,
) C.int32_t {
	return C.int32_t(useAbilityRewardLegacy53FAE0(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(item)),
	))
}
