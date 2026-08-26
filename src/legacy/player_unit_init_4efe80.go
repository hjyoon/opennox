package legacy

/*
#include <stdint.h>

uint32_t* sub_56F920(int token, int delta);

static inline void nox_playerUnitInit_protectGold_4EFE80(
		uint32_t token, int32_t delta) {
	(void)sub_56F920((int32_t)token, delta);
}
*/
import "C"

import (
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

func playerUnitInitProtectGold4EFE80(token uint32, delta int32) {
	C.nox_playerUnitInit_protectGold_4EFE80(
		C.uint32_t(token),
		C.int32_t(delta),
	)
}

// Nox_xxx_protectGoldDelta_56F920 applies the scalar protection update used
// by 004FA590/004FA5D0 without passing a native Object pointer through their
// original int-typed ABI.
func Nox_xxx_protectGoldDelta_56F920(token uint32, delta int32) {
	playerUnitInitProtectGold4EFE80(token, delta)
}

func playerUnitInitRuntime4EFE80() server.PlayerUnitInitRuntime4EFE80 {
	s := GetServer().S()
	return server.PlayerUnitInitRuntime4EFE80{
		ProtectGold: playerUnitInitProtectGold4EFE80,
		SyncLevel: func(unit *server.Object) {
			_ = playerSyncLevelCall4EF140(unit)
		},
		AwardBeastScrolls: Nox_xxx_spellAwardAll1_4EFD80,
		AwardSpells:       Nox_xxx_spellAwardAll2_4EFC80,
		ReadValues: func(unit *server.Object, rewardArg int32) {
			_ = playerReadValuesCall4EEDC0(unit, rewardArg)
		},
		AwardWarriorAbilities: Nox_xxx_spellAwardAll3_4EFE10,
		GameFlag: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		BalanceFloat: func(key string) float32 {
			// 004EFE80 spills the x87 result to a binary32 stack slot before
			// calling nox_float2int.
			return float32(s.Balance.Float(key))
		},
		MakeDefaultItems: func(unit *server.Object, restoreStats, keepItems int32) uint8 {
			return playerMakeDefItemsCall4EF7D0(unit, restoreStats, keepItems)
		},
	}
}

func playerUnitInitCall4EFE80(unit *server.Object) uint8 {
	return GetServer().S().PlayerUnitInit4EFE80(
		unit,
		playerUnitInitRuntime4EFE80(),
	)
}
