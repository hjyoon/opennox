package server

import (
	"math"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

const (
	monsterSetFightTargetClassLow515D30 = uint8(0x02)
	monsterSetFightTargetDeadFlag515D30 = uint32(0x8000)
)

// The PE32 routine caches the monster update-data pointer before checking the
// target, then reads the target position only after both action pushes.
// Handles remain native-width; position, frame, and action values are dwords.
type monsterSetFightTargetHooks515D30[O, U, A comparable] struct {
	loadUpdate         func(O) U
	loadClassLow       func(O) uint8
	loadFlags          func(O) uint32
	clearActionStack   func(O)
	storeTarget        func(U, O)
	setNextFrame       func()
	pushAction         func(O, uint32) A
	storeArgBits       func(A, int, uint32)
	loadTargetPosXBits func(O) uint32
	loadTargetPosYBits func(O) uint32
	loadFrame          func() uint32
}

// monsterSetFightTarget515D30 restores GAME.EXE 00515D30. An absent update
// pointer is rejected instead of reproducing the original invalid dereference.
func monsterSetFightTarget515D30[O, U, A comparable](unit, target O, h monsterSetFightTargetHooks515D30[O, U, A]) {
	var nilObject O
	if unit == nilObject {
		return
	}
	update := h.loadUpdate(unit)
	if target == nilObject || h.loadClassLow(unit)&monsterSetFightTargetClassLow515D30 == 0 ||
		unit == target || h.loadFlags(unit)&monsterSetFightTargetDeadFlag515D30 != 0 {
		return
	}
	var nilUpdate U
	if update == nilUpdate {
		return
	}
	h.clearActionStack(unit)
	h.storeTarget(update, target)
	h.setNextFrame()
	var nilAction A
	if item := h.pushAction(unit, uint32(ai.ACTION_REPORT)); item != nilAction {
		h.storeArgBits(item, 0, uint32(ai.ACTION_FIGHT))
	}
	if item := h.pushAction(unit, uint32(ai.ACTION_FIGHT)); item != nilAction {
		h.storeArgBits(item, 0, h.loadTargetPosXBits(target))
		h.storeArgBits(item, 1, h.loadTargetPosYBits(target))
		h.storeArgBits(item, 2, h.loadFrame())
	}
}

// MonsterSetFightTarget515D30 replaces the script Attack(object) C body,
// which truncated both object pointers to signed 32-bit integers.
func (s *Server) MonsterSetFightTarget515D30(unit, target *Object) {
	monsterSetFightTarget515D30(unit, target, monsterSetFightTargetHooks515D30[*Object, *MonsterUpdateData, *AIStackItem]{
		loadUpdate: func(unit *Object) *MonsterUpdateData {
			if unit.UpdateData == nil {
				return nil
			}
			return unit.UpdateDataMonster()
		},
		loadClassLow: func(unit *Object) uint8 { return uint8(unit.ObjClass) },
		loadFlags:    func(unit *Object) uint32 { return uint32(unit.ObjFlags) },
		clearActionStack: func(unit *Object) {
			unit.ClearActionStack()
		},
		storeTarget: func(update *MonsterUpdateData, target *Object) {
			update.PreferredEnemy = target
		},
		setNextFrame: func() {
			*memmap.PtrUint32(0x5D4594, 2487684) = s.Frame() + 1
		},
		pushAction: func(unit *Object, action uint32) *AIStackItem {
			return unit.MonsterPushAction(ai.ActionType(action))
		},
		storeArgBits: func(item *AIStackItem, index int, bits uint32) {
			item.Args[index] = uintptr(bits)
		},
		loadTargetPosXBits: func(target *Object) uint32 {
			return math.Float32bits(target.PosVec.X)
		},
		loadTargetPosYBits: func(target *Object) uint32 {
			return math.Float32bits(target.PosVec.Y)
		},
		loadFrame: s.Frame,
	})
}
