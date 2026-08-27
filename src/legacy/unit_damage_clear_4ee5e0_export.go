package legacy

/*
#include <stdint.h>
#include "unit_damage_clear_4ee5e0.h"

int nox_xxx_monsterCallDieFn_50A3D0(uint32_t* unit);
*/
import "C"

import (
	"log/slog"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/server"
)

func unitDamageClearMonsterDie4EE5E0(unit *server.Object) {
	s := GetServer().S()
	var unsupportedReason string
	runtime := server.MonsterDieRuntime50A3D0{
		GameFlag: func(flag uint32) bool {
			return noxflags.HasGame(noxflags.GameFlag(flag))
		},
		IsZombie:     s.IsZombie,
		ObserveClear: Nox_xxx_playerObserveClear_4DDEF0,
		RemoveShadow: Nox_xxx_action_4DA9F0,
		RandomInt: func(minimum, maximum int) int {
			if s.Rand.Logic == nil {
				return minimum
			}
			return s.Rand.Logic.IntClamp(minimum, maximum)
		},
		SetDecayTime: func(obj *server.Object, frames uint32) {
			s.DecaySetTime511660(obj, frames)
		},
		NetFxShield: func(index int, obj *server.Object) {
			s.Nox_xxx_netFxShield_0_4D9200(index, obj)
		},
		UnmarkMinimap: func(index int, obj *server.Object, flags uint32) {
			s.Players.Nox_xxx_netUnmarkMinimapObj_417300(ntype.PlayerInd(index), obj, flags)
		},
		DropAllItems: func(obj *server.Object) {
			dropAllItemsCall4EDA40(obj)
		},
		Unsupported: func(reason string, _ *server.Object) {
			unsupportedReason = reason
		},
	}
	if Sub_4FC0B0 != nil {
		runtime.AwardSoloKill = func(killer *server.Object) {
			Sub_4FC0B0(killer, 1)
		}
	}
	if s.MonsterDieNative50A3D0(unit, runtime) {
		return
	}

	// The original dispatcher is ABI-safe on a native 32-bit build. On a
	// 64-bit build its uint32 pointer fields truncate Object and Player
	// addresses, so an unsupported branch must never be sent back through C.
	if unsafe.Sizeof(uintptr(0)) == 4 {
		C.nox_xxx_monsterCallDieFn_50A3D0((*C.uint32_t)(unit.CObj()))
		return
	}
	if s.Log != nil {
		s.Log.Error("MonsterDie native branch is not ported",
			slog.String("reason", unsupportedReason),
			slog.Uint64("unit_ptr", uint64(uintptr(unit.CObj()))),
		)
	}
	GetServer().DelayedDelete(unit)
}

//export nox_xxx_unitDamageClear_4EE5E0
func nox_xxx_unitDamageClear_4EE5E0(unit *C.nox_object_t, damage C.int) {
	unitDamageClearCall4EE5E0(
		asObjectS((*nox_object_t)(unit)),
		int32(damage),
	)
}
