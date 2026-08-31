package legacy

/*
#include <stdint.h>

#include "ability_grant_4eed40.h"

int nox_xxx_isQuest_4D6F50(void);
int sub_4D6F70(void);
*/
import "C"

import (
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

const abilityGivePlayerAllTableOffset4EED40 = uintptr(206108)

func abilityGivePlayerAllRuntime4EED40() server.AbilityGivePlayerAllRuntime4EED40 {
	return server.AbilityGivePlayerAllRuntime4EED40{
		LoadAbilityID: func(index int32) uint32 {
			return *memmap.PtrUint32(
				0x587000,
				abilityGivePlayerAllTableOffset4EED40+uintptr(index)*4,
			)
		},
		IsQuest: func() int32 {
			return int32(C.nox_xxx_isQuest_4D6F50())
		},
		QuestMode: func() int32 {
			return int32(C.sub_4D6F70())
		},
		RewardAbility: func(unit *server.Object, ability, rewardArg int32) {
			_ = abilityRewardServLegacy4FB9C0(unit, ability, rewardArg)
		},
	}
}

func abilityGivePlayerAllCall4EED40(unit *server.Object, count int8, rewardArg int32) {
	GetServer().S().AbilityGivePlayerAll4EED40(
		unit,
		count,
		rewardArg,
		abilityGivePlayerAllRuntime4EED40(),
	)
}

// Nox_xxx_abilGivePlayerAll_4EED40 gives Go-owned callers the restored
// native-pointer path.
func Nox_xxx_abilGivePlayerAll_4EED40(unit *server.Object, count int8, rewardArg int32) {
	abilityGivePlayerAllCall4EED40(unit, count, rewardArg)
}
