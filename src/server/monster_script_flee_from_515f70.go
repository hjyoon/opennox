package server

import (
	"math"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

const (
	monsterScriptFleeClassLow515F70 = uint8(0x02)
	monsterScriptFleeDeadFlag515F70 = uint32(0x8000)
)

// The PE32 routine pushes REPORT, DEPENDENCY_TIME, then FLEE. It reads the
// target position only after the final push; failed pushes do not roll back
// earlier actions. Object and update handles must retain their native width.
type monsterScriptFleeHooks515F70[O, U, A comparable] struct {
	loadClassLow       func(O) uint8
	loadFlags          func(O) uint32
	loadUpdate         func(O) U
	hasAction          func(U, uint32) bool
	pushAction         func(O, uint32) A
	loadFrame          func() uint32
	loadTargetPosXBits func(O) uint32
	loadTargetPosYBits func(O) uint32
	storeArgBits       func(A, int, uint32)
}

// monsterScriptFleeFrom515F70 restores GAME.EXE 00515F70. A nil target or
// update pointer is rejected safely instead of reproducing the PE32 fault.
func monsterScriptFleeFrom515F70[O, U, A comparable](unit, target O, duration uint32, h monsterScriptFleeHooks515F70[O, U, A]) {
	var nilObject O
	if unit == nilObject || h.loadClassLow(unit)&monsterScriptFleeClassLow515F70 == 0 ||
		h.loadFlags(unit)&monsterScriptFleeDeadFlag515F70 != 0 {
		return
	}
	update := h.loadUpdate(unit)
	var nilUpdate U
	if update == nilUpdate || h.hasAction(update, uint32(ai.ACTION_FLEE)) || target == nilObject {
		return
	}
	var nilAction A
	if item := h.pushAction(unit, uint32(ai.ACTION_REPORT)); item != nilAction {
		h.storeArgBits(item, 0, uint32(ai.ACTION_FLEE))
	}
	if item := h.pushAction(unit, uint32(ai.DEPENDENCY_TIME)); item != nilAction {
		h.storeArgBits(item, 0, h.loadFrame()+duration)
	}
	if item := h.pushAction(unit, uint32(ai.ACTION_FLEE)); item != nilAction {
		h.storeArgBits(item, 0, h.loadTargetPosXBits(target))
		posY := h.loadTargetPosYBits(target)
		h.storeArgBits(item, 2, 0)
		h.storeArgBits(item, 1, posY)
	}
}

// MonsterScriptFleeFrom515F70 replaces the script Flee C body, which assumed
// PE32 object pointers and a two-dword {target, duration} argument record.
func (s *Server) MonsterScriptFleeFrom515F70(unit, target *Object, duration int) {
	monsterScriptFleeFrom515F70(unit, target, uint32(int32(duration)), monsterScriptFleeHooks515F70[*Object, *MonsterUpdateData, *AIStackItem]{
		loadClassLow: func(unit *Object) uint8 { return uint8(unit.ObjClass) },
		loadFlags:    func(unit *Object) uint32 { return uint32(unit.ObjFlags) },
		loadUpdate: func(unit *Object) *MonsterUpdateData {
			if unit.UpdateData == nil {
				return nil
			}
			return unit.UpdateDataMonster()
		},
		hasAction: func(update *MonsterUpdateData, action uint32) bool {
			return update.HasAction(ai.ActionType(action))
		},
		pushAction: func(unit *Object, action uint32) *AIStackItem {
			return unit.MonsterPushAction(ai.ActionType(action))
		},
		loadFrame: s.Frame,
		loadTargetPosXBits: func(target *Object) uint32 {
			return math.Float32bits(target.PosVec.X)
		},
		loadTargetPosYBits: func(target *Object) uint32 {
			return math.Float32bits(target.PosVec.Y)
		},
		storeArgBits: func(item *AIStackItem, index int, bits uint32) {
			item.Args[index] = uintptr(bits)
		},
	})
}
