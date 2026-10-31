package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
_Static_assert(_Generic(&nox_xxx_mobActionMelee1_532130, uintptr_t (*)(nox_object_t*): 1, default: 0), "native melee-start ABI");
_Static_assert(_Generic(&sub_532390, void (*)(nox_object_t*, nox_object_t*): 1, default: 0), "native blocker ABI");
_Static_assert(_Generic(&nox_xxx_mobActionMeleeAtt_532440, char (*)(nox_object_t*): 1, default: 0), "native melee-update ABI");
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func monsterMeleeNativeRuntime532130(binding Server) server.MonsterMeleeNativeRuntime532130 {
	ordinary := monsterActionMeleeRuntime532130()
	runtime := server.MonsterMeleeNativeRuntime532130{
		AudioEvent: ordinary.AudioEvent, BuffOff: ordinary.BuffOff, PlayerAttack: ordinary.PlayerAttack,
		CanStrike: ordinary.CanStrike, Strike: ordinary.Strike,
	}
	if provider, ok := binding.(interface {
		MonsterMeleeNativeRuntime532130() server.MonsterMeleeNativeRuntime532130
	}); ok {
		provided := provider.MonsterMeleeNativeRuntime532130()
		if provided.AudioEvent == nil {
			provided.AudioEvent = runtime.AudioEvent
		}
		if provided.BuffOff == nil {
			provided.BuffOff = runtime.BuffOff
		}
		if provided.PlayerAttack == nil {
			provided.PlayerAttack = runtime.PlayerAttack
		}
		if provided.CanStrike == nil {
			provided.CanStrike = runtime.CanStrike
		}
		if provided.Strike == nil {
			provided.Strike = runtime.Strike
		}
		runtime = provided
	}
	return runtime
}

//export nox_xxx_mobActionMelee1_532130
func nox_xxx_mobActionMelee1_532130(unit *C.nox_object_t) C.uintptr_t {
	binding := GetServer()
	return C.uintptr_t(binding.S().MonsterMeleeNative532130(0x532130, asObjectS((*nox_object_t)(unit)), server.MonsterMeleeGlobalState532390(), monsterMeleeNativeRuntime532130(binding)))
}

//export sub_532390
func sub_532390(candidate, unit *C.nox_object_t) {
	server.MonsterMeleeBlocker532390(asObjectS((*nox_object_t)(candidate)), asObjectS((*nox_object_t)(unit)), server.MonsterMeleeGlobalState532390())
}

//export nox_xxx_mobActionMeleeAtt_532440
func nox_xxx_mobActionMeleeAtt_532440(unit *C.nox_object_t) C.char {
	binding := GetServer()
	return C.char(binding.S().MonsterMeleeNative532130(0x532440, asObjectS((*nox_object_t)(unit)), server.MonsterMeleeGlobalState532390(), monsterMeleeNativeRuntime532130(binding)))
}
