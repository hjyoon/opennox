package server

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

const monsterScriptPauseDeadFlag516090 = uint32(0x8000)

// Native-width object and action handles are kept separate from the PE32
// dwords stored in AI action arguments.
type monsterScriptPauseHooks516090[O, A comparable] struct {
	loadClassLow func(O) uint8
	loadFlags    func(O) uint32
	pushAction   func(O, uint32) A
	loadFrame    func() uint32
	storeArgBits func(A, int, uint32)
}

// monsterScriptPause516090 restores GAME.EXE 00516090. The original attempts
// REPORT first, then independently attempts WAIT even if REPORT could not be
// pushed. It reads the current frame only after a successful WAIT push.
func monsterScriptPause516090[O, A comparable](unit O, duration uint32, h monsterScriptPauseHooks516090[O, A]) {
	var nilUnit O
	if unit == nilUnit || h.loadClassLow(unit)&uint8(object.ClassMonster) == 0 ||
		h.loadFlags(unit)&monsterScriptPauseDeadFlag516090 != 0 {
		return
	}

	var nilAction A
	if item := h.pushAction(unit, uint32(ai.ACTION_REPORT)); item != nilAction {
		h.storeArgBits(item, 0, uint32(ai.ACTION_WAIT))
	}
	if item := h.pushAction(unit, uint32(ai.ACTION_WAIT)); item != nilAction {
		h.storeArgBits(item, 0, h.loadFrame()+duration)
	}
}

// MonsterScriptPause516090 replaces the legacy C implementation, which
// truncated the object pointer to a PE32 int before reading it.
func (s *Server) MonsterScriptPause516090(unit *Object, duration int) {
	if unit == nil || unit.UpdateData == nil {
		return
	}
	monsterScriptPause516090(unit, uint32(int32(duration)), monsterScriptPauseHooks516090[*Object, *AIStackItem]{
		loadClassLow: func(unit *Object) uint8 { return uint8(unit.ObjClass) },
		loadFlags:    func(unit *Object) uint32 { return uint32(unit.ObjFlags) },
		pushAction: func(unit *Object, action uint32) *AIStackItem {
			return unit.MonsterPushAction(ai.ActionType(action))
		},
		loadFrame: s.Frame,
		storeArgBits: func(item *AIStackItem, index int, bits uint32) {
			item.Args[index] = uintptr(bits)
		},
	})
}
