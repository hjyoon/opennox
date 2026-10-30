package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
_Static_assert(sizeof(uintptr_t) == sizeof(void*), "scheduler result is a native word");
#define NOX_FIGHT_SCHEDULER_ABI(fn) _Static_assert(_Generic(&(fn), int* (*)(nox_object_t*, nox_object_t*): 1, default: 0), #fn " native unit/target ABI")
NOX_FIGHT_SCHEDULER_ABI(nox_xxx_TODOsomeCallingMeleeAttack_531B40);
NOX_FIGHT_SCHEDULER_ABI(sub_531C10);
NOX_FIGHT_SCHEDULER_ABI(nox_xxx_monsterAction_531C60);
NOX_FIGHT_SCHEDULER_ABI(sub_531D50);
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func monsterFightSchedulerRuntime531B40() server.MonsterFightSchedulerRuntime531B40 {
	return server.MonsterFightSchedulerRuntime531B40{
		CanSummon: func(unit *server.Object, guideIndex int) bool {
			return Nox_xxx_checkSummonedCreaturesLimit_500D70 != nil &&
				Nox_xxx_checkSummonedCreaturesLimit_500D70(unit, int32(guideIndex))
		},
	}
}

//export nox_monster_fight_scheduler_native_531B40
func nox_monster_fight_scheduler_native_531B40(unit, target *C.nox_object_t, entry C.uint) C.uintptr_t {
	binding := GetServer()
	runtime := monsterFightSchedulerRuntime531B40()
	if provider, ok := binding.(interface {
		MonsterFightSchedulerRuntime531B40() server.MonsterFightSchedulerRuntime531B40
	}); ok {
		provided := provider.MonsterFightSchedulerRuntime531B40()
		if provided.CanSummon == nil {
			provided.CanSummon = runtime.CanSummon
		}
		runtime = provided
	}
	return C.uintptr_t(binding.S().MonsterFightScheduler531B40(uint32(entry), asObjectS((*nox_object_t)(unit)), asObjectS((*nox_object_t)(target)), runtime))
}
