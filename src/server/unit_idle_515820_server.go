package server

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func unitIdleNative515820(unit *Object) {
	unitIdle515820(unit, unitIdleHooks515820[*Object]{
		loadClassLow: func(unit *Object) uint8 {
			return uint8(unit.ObjClass)
		},
		loadFlags: func(unit *Object) uint32 {
			return uint32(unit.ObjFlags)
		},
		clearActionStack: func(unit *Object) {
			unit.ClearActionStack()
		},
		pushAction: func(unit *Object, action uint32) {
			unit.MonsterPushAction(ai.ActionType(action))
		},
	})
}

// UnitIdle515820 binds GAME.EXE 00515820 to a native-width Object pointer.
// It deliberately adds no UpdateData guard because the original routine calls
// both action-stack functions after only its null, class, and flag gates.
func (*Server) UnitIdle515820(unit *Object) {
	unitIdleNative515820(unit)
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.ObjClass)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.ObjFlags)]
)
