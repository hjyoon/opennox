package server

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// MonsterLookAt5125A0 keeps native Object/AIStackItem identities and the
// original initial-load fault for a null object. No server or update-data
// lookup is needed until both original gates and the coordinate spills pass.
func MonsterLookAt5125A0(unit *Object, direction int32) uintptr {
	return monsterLookAt5125A0(unit, direction, monsterLookAtHooks5125A0[*Object, *AIStackItem]{
		directionToAngle: DirectionIndexToAngle509E90,
		loadClassLow:     func(unit *Object) uint8 { return uint8(unit.ObjClass) },
		loadFlags:        func(unit *Object) uint32 { return uint32(unit.ObjFlags) },
		loadCosine: func(angle uint32) float32 {
			cosine, _ := SinCosDir(byte(angle))
			return cosine
		},
		loadSine: func(angle uint32) float32 {
			_, sine := SinCosDir(byte(angle))
			return sine
		},
		loadX: func(unit *Object) float32 { return unit.PosVec.X },
		loadY: func(unit *Object) float32 { return unit.PosVec.Y },
		pushAction: func(unit *Object, action uint32) *AIStackItem {
			// Do not set variadic arguments before PushAction returns: it may
			// reject a full/dead stack or cancel an existing action first.
			return unit.MonsterPushActionImpl(ai.ActionType(action), "go", 0)
		},
		storeArgBits: func(action *AIStackItem, index int, bits uint32) {
			action.Args[index] = uintptr(bits)
		},
		actionResult: func(action *AIStackItem) uintptr { return uintptr(unsafe.Pointer(action)) },
	})
}
