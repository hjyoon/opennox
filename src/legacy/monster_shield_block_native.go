package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func monsterTestBlockShield533E70(unit *server.Object) int {
	found := memmap.PtrUint32(0x5D4594, 2487956)
	*found = 0
	// The radius is 100. GAME.EXE 00533E8D seeds the best distance with
	// different bits (0x4e6e6b28, or 1e9), admitting the radius boundary.
	*memmap.PtrUint32(0x5D4594, 2487988) = 0x4e6e6b28
	GetServer().S().Map.EachMissileInCirclePtr518170(&unit.PosVec, 100, func(missile *server.Object) bool {
		considerShieldBlockMissile533EB0(unit, missile)
		return true
	})
	return int(*found)
}

//export nox_xxx_monsterTestBlockShield_533E70
func nox_xxx_monsterTestBlockShield_533E70(unit *nox_object_t) C.int {
	return C.int(monsterTestBlockShield533E70(asObjectS(unit)))
}

//export sub_533EB0
func sub_533EB0(missile *nox_object_t, unit unsafe.Pointer) {
	considerShieldBlockMissile533EB0(asObjectS((*nox_object_t)(unit)), asObjectS(missile))
}

func considerShieldBlockMissile533EB0(unit, missile *server.Object) {
	previous := missile.PrevPos
	if sub_54E6F0(asObjectC(unit), asObjectC(missile)) == 0 {
		return
	}
	dir := int32(int16(unit.Direction1))
	if twoPointsAndDirection4E6E50(unit.PosVec, dir, previous)&1 == 0 {
		return
	}
	direction := *(*types.Pointf)(memmap.PtrOff(0x587000, uintptr(194136+8*dir)))
	distance, incoming := server.ShieldBlockApproach533EB0(unit.PosVec, missile.PosVec, missile.VelVec, direction)
	if !incoming || !GetServer().S().CanInteract(unit, missile, 0) {
		return
	}
	best := memmap.PtrFloat32(0x5D4594, 2487988)
	if !(distance >= *best) {
		*memmap.PtrUint32(0x5D4594, 2487956) = 1
		*best = distance
	}
}
