package server

import (
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func monsterWalkToNative514110(unit *Object, x, y float32) {
	monsterWalkTo514110(unit, math.Float32bits(x), math.Float32bits(y), monsterWalkToHooks514110[*Object, *AIStackItem]{
		loadFlags: func(unit *Object) uint32 {
			return uint32(unit.ObjFlags)
		},
		loadClassLow: func(unit *Object) uint8 {
			return uint8(unit.ObjClass)
		},
		clearActionStack: func(unit *Object) {
			unit.ClearActionStack()
		},
		pushAction: func(unit *Object, action uint32) *AIStackItem {
			return unit.MonsterPushAction(ai.ActionType(action))
		},
		storeArgBits: func(action *AIStackItem, index int, bits uint32) {
			action.Args[index] = uintptr(bits)
		},
	})
}

// MonsterWalkTo514110 binds GAME.EXE 00514110 to native-width Object and
// AIStackItem pointers. The original routine intentionally faults for a null
// object before any gate, so this adapter does not introduce a nil guard.
func (*Server) MonsterWalkTo514110(unit *Object, x, y float32) {
	monsterWalkToNative514110(unit, x, y)
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.ObjClass)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.ObjFlags)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(AIStackItem{}.Action)]
)
