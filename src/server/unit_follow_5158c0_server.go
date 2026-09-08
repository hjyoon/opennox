package server

import (
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func unitFollowNative5158C0(unit, target *Object) {
	unitFollow5158C0(unit, target, unitFollowHooks5158C0[*Object, *AIStackItem]{
		loadClassLow: func(unit *Object) uint8 {
			return uint8(unit.ObjClass)
		},
		loadFlags: func(unit *Object) uint32 {
			return uint32(unit.ObjFlags)
		},
		clearActionStack: func(unit *Object) {
			unit.ClearActionStack()
		},
		pushAction: func(unit *Object, action uint32) *AIStackItem {
			return unit.MonsterPushAction(ai.ActionType(action))
		},
		loadPosXBits: func(target *Object) uint32 {
			return math.Float32bits(target.PosVec.X)
		},
		storeArgBits: func(action *AIStackItem, index int, bits uint32) {
			action.Args[index] = uintptr(bits)
		},
		loadPosYBits: func(target *Object) uint32 {
			return math.Float32bits(target.PosVec.Y)
		},
		storeTarget: func(action *AIStackItem, index int, target *Object) {
			action.Args[index] = uintptr(unsafe.Pointer(target))
		},
	})
}

// UnitSetFollow5158C0 binds GAME.EXE 005158C0 to native-width Object and
// AIStackItem pointers. It adds no UpdateData guard because the original only
// applies its null, class, identity, and flag gates before stack access.
func (*Server) UnitSetFollow5158C0(unit, target *Object) {
	unitFollowNative5158C0(unit, target)
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.ObjClass)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.ObjFlags)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.PosVec.X)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.PosVec.Y)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(AIStackItem{}.Action)]
)
