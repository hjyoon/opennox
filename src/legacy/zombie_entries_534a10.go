package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

var zombieEntryTypes534A10 = func() server.TypePredicateRuntime534A10 { return GetServer().S().TypePredicateRuntime534A10() }
var zombieEntryPop534A90 = func(unit *server.Object) uint32 { return uint32(unit.MonsterPopAction()) }
var zombieEntryRaise534AB0 = func() server.MonsterRaiseZombieRawRuntime534AB0 {
	return server.MonsterRaiseZombieRawRuntime534AB0{
		IsZombie: func(unit *server.Object) uint32 { return server.UnitIsZombie534A40(unit, zombieEntryTypes534A10()) },
		CurrentAction: func(unit *server.Object) uint32 {
			return uint32((*server.MonsterUpdateData)(unit.UpdateData).AIStackHead().Type())
		},
		Pop: zombieEntryPop534A90,
		Push: func(unit *server.Object, action uint32) uintptr {
			return uintptr(unsafe.Pointer(unit.MonsterPushAction(ai.ActionType(action))))
		},
		Audio:  func(id uint32, unit *server.Object) { GetServer().S().Audio.EventObj(sound.ID(id), unit, 0, 0) },
		Health: Nox_xxx_unitHPsetOnMax_4EE6F0,
	}
}

//export nox_unit_is_plant_native_534A10
func nox_unit_is_plant_native_534A10(unit *nox_object_t) C.int {
	return C.int(server.UnitIsPlant534A10(asObjectS(unit), zombieEntryTypes534A10()))
}

//export nox_unit_is_zombie_native_534A40
func nox_unit_is_zombie_native_534A40(unit *nox_object_t) C.int {
	return C.int(server.UnitIsZombie534A40(asObjectS(unit), zombieEntryTypes534A10()))
}

//export nox_mob_get_up_native_534A90
func nox_mob_get_up_native_534A90(unit *nox_object_t) C.char {
	return C.char(server.MonsterActionGetUpRaw534A90(asObjectS(unit), zombieEntryPop534A90))
}

//export nox_mob_raise_zombie_native_534AB0
func nox_mob_raise_zombie_native_534AB0(unit *nox_object_t) C.uint32_t {
	return C.uint32_t(server.MonsterRaiseZombieRaw534AB0(asObjectS(unit), zombieEntryRaise534AB0()))
}
