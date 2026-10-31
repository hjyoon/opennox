package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
_Static_assert(_Generic(&sub_532540, uintptr_t (*)(nox_object_t*): 1, default: 0), "native missile-start ABI");
_Static_assert(_Generic(&nox_xxx_mobActionMissileAtt_532610, char (*)(nox_object_t*): 1, default: 0), "native missile-update ABI");
_Static_assert(_Generic(&nox_xxx_projAddVelocitySmth_533080, uintptr_t (*)(nox_object_t*, nox_object_t*, float, float2*): 1, default: 0), "native missile-prediction ABI");
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func monsterMissileNativeRuntime532540(binding Server) server.MonsterMissileNativeRuntime532540 {
	ordinary := monsterActionMissileRuntime532540()
	runtime := server.MonsterMissileNativeRuntime532540{
		AudioEvent: ordinary.AudioEvent, BuffOff: ordinary.BuffOff, PlayerAttack: ordinary.PlayerAttack,
		CreateObjectAt: ordinary.CreateObjectAt, DelayedDelete: ordinary.DelayedDelete,
	}
	if provider, ok := binding.(interface {
		MonsterMissileNativeRuntime532540() server.MonsterMissileNativeRuntime532540
	}); ok {
		provided := provider.MonsterMissileNativeRuntime532540()
		if provided.AudioEvent == nil {
			provided.AudioEvent = runtime.AudioEvent
		}
		if provided.BuffOff == nil {
			provided.BuffOff = runtime.BuffOff
		}
		if provided.PlayerAttack == nil {
			provided.PlayerAttack = runtime.PlayerAttack
		}
		if provided.CreateObjectAt == nil {
			provided.CreateObjectAt = runtime.CreateObjectAt
		}
		if provided.DelayedDelete == nil {
			provided.DelayedDelete = runtime.DelayedDelete
		}
		runtime = provided
	}
	return runtime
}

//export sub_532540
func sub_532540(unit *C.nox_object_t) C.uintptr_t {
	binding := GetServer()
	return C.uintptr_t(binding.S().MonsterMissileNative532540(0x532540, asObjectS((*nox_object_t)(unit)), monsterMissileNativeRuntime532540(binding)))
}

//export nox_xxx_mobActionMissileAtt_532610
func nox_xxx_mobActionMissileAtt_532610(unit *C.nox_object_t) C.char {
	binding := GetServer()
	return C.char(binding.S().MonsterMissileNative532540(0x532610, asObjectS((*nox_object_t)(unit)), monsterMissileNativeRuntime532540(binding)))
}

//export nox_xxx_projAddVelocitySmth_533080
func nox_xxx_projAddVelocitySmth_533080(unit, target *C.nox_object_t, speed C.float, output *C.float2) C.uintptr_t {
	result := server.MonsterProjectileLead533080(asObjectS((*nox_object_t)(unit)), asObjectS((*nox_object_t)(target)), float32(speed), (*types.Pointf)(unsafe.Pointer(output)))
	return C.uintptr_t(uintptr(unsafe.Pointer(result)))
}
