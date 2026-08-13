package server

import (
	"github.com/opennox/libs/object"
)

type unitIsHostileMimicHooks4E7F90[O comparable] struct {
	loadMimicCache  func() uint32
	lookupType      func(string) uint32
	storeMimicCache func(uint32)
	isEnemy         func(O, O) int32
	isQuest         func() int32
	loadType        func(O) uint16
	loadClassLow    func(O) uint8
	loadOwner       func(O) O
}

// unitIsHostileMimic4E7F90 preserves GAME.EXE 004E7F90. Mimic cache
// initialization precedes both null checks. The cache is loaded again after
// the enemy and Quest callbacks, and only the complete Quest exception clears
// the base result produced by an exact zero enemy result.
func unitIsHostileMimic4E7F90[O comparable](
	obj, obj2 O,
	hooks unitIsHostileMimicHooks4E7F90[O],
) int32 {
	mimicType := hooks.loadMimicCache()
	if mimicType == 0 {
		mimicType = hooks.lookupType("Mimic")
		hooks.storeMimicCache(mimicType)
	}

	var zero O
	if obj == zero || obj2 == zero {
		return 0
	}

	var result int32
	if hooks.isEnemy(obj, obj2) == 0 {
		result = 1
	}
	if hooks.isQuest() == 0 {
		return result
	}

	mimicType = hooks.loadMimicCache()
	if uint32(hooks.loadType(obj2)) != mimicType {
		return result
	}
	if hooks.loadClassLow(obj)&uint8(object.ClassPlayer) == 0 {
		return result
	}
	if hooks.loadOwner(obj2) != zero {
		return result
	}
	return 0
}
