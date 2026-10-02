package server

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// UnitHunt5157A0 preserves the original null/class/dead gates and the two
// action-stack calls, using native fields instead of PE32 pointer offsets.
// A rejected object needs no active server or UpdateData.
func UnitHunt5157A0(unit *Object) {
	unitHunt5157A0(unit, unitHuntHooks5157A0[*Object]{
		loadClassLow: func(unit *Object) uint8 { return uint8(unit.ObjClass) },
		loadFlags:    func(unit *Object) uint32 { return uint32(unit.ObjFlags) },
		clearActionStack: func(unit *Object) {
			unit.ClearActionStack()
		},
		pushAction: func(unit *Object, action uint32) {
			unit.MonsterPushAction(ai.ActionType(action))
		},
	})
}

func (*Server) UnitHunt5157A0(unit *Object) {
	UnitHunt5157A0(unit)
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.ObjClass)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.ObjFlags)]
)
