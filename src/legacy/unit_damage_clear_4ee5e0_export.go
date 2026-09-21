package legacy

/*
#include <stdint.h>
#include "unit_damage_clear_4ee5e0.h"

int nox_xxx_monsterCallDieFn_50A3D0(nox_object_t* unit);
*/
import "C"

import (
	"log/slog"

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
		QuestPrepare: s.MonsterSpawnCleanupUnlessZombie50E1E0,
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
		AwardSoloKill: func(killer *server.Object) {
			Sub_4FC0B0(killer, 1)
		},
		CreditQuestKill: func(killer *server.Object) {
			killer.RecordMonsterKilled4D6170()
		},
		Unsupported: func(reason string, _ *server.Object) {
			unsupportedReason = reason
		},
	}
	if s.MonsterDieNative50A3D0(unit, runtime) {
		return
	}

	if s.Log != nil {
		s.Log.Error("MonsterDie native dispatcher rejected its input",
			slog.String("reason", unsupportedReason),
			slog.Uint64("unit_ptr", uint64(uintptr(unit.CObj()))),
		)
	}
	GetServer().DelayedDelete(unit)
}

var monsterDieExportImpl50A3D0 = func(unit *server.Object) int32 {
	unitDamageClearMonsterDie4EE5E0(unit)
	return 1
}

func monsterDieExportCall50A3D0(unit *server.Object) int32 {
	return int32(C.nox_xxx_monsterCallDieFn_50A3D0(asObjectC(unit)))
}

//export nox_xxx_monsterCallDieFn_50A3D0
func nox_xxx_monsterCallDieFn_50A3D0(unit *C.nox_object_t) C.int {
	return C.int(monsterDieExportImpl50A3D0(asObjectS((*nox_object_t)(unit))))
}

//export nox_xxx_unitDamageClear_4EE5E0
func nox_xxx_unitDamageClear_4EE5E0(unit *C.nox_object_t, damage C.int) {
	unitDamageClearCall4EE5E0(
		asObjectS((*nox_object_t)(unit)),
		int32(damage),
	)
}
