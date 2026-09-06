package server

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// MonsterActionBlockRuntime532070 supplies the remaining shield query above
// the server package without narrowing the unit pointer to the PE32 ABI.
type MonsterActionBlockRuntime532070 struct {
	TestShield func(*Object) int
}

type monsterActionBlockHooks532070 struct {
	frame      func() uint32
	tickRate   func() uint32
	testShield func(*Object) int
	pop        func() int
	push       func(ai.ActionType, ...any) *AIStackItem
}

// monsterActionBlockAttack532070 restores GAME.EXE 00532070. The current
// stack item is intentionally cached across TestShield, while the subclass is
// read after Pop: both details are observable in the original instruction
// order. Frame arithmetic and expiry comparison retain uint32 wrap semantics.
func monsterActionBlockAttack532070(unit *Object, hooks monsterActionBlockHooks532070) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) ||
		hooks.frame == nil || hooks.tickRate == nil || hooks.testShield == nil || hooks.pop == nil || hooks.push == nil {
		return false
	}
	update := unit.UpdateDataMonster()
	index := int(update.AIStackInd)
	if index < 0 || index >= len(update.AIStack) {
		return false
	}
	head := &update.AIStack[index]
	if hooks.testShield(unit) != 0 {
		halfRate := hooks.tickRate() >> 1
		head.Args[0] = uintptr(hooks.frame() + halfRate)
	}
	frame := hooks.frame()
	deadline := head.ArgU32(0)
	if frame > deadline {
		hooks.pop()
		if !unit.ObjSubClass.AsMonster().Has(object.MonsterNPC) {
			hooks.push(ai.ACTION_BLOCK_FINISH)
		}
	}
	return true
}

// MonsterActionBlockAttack532070 binds the restored shield-block action to
// the live frame clock and native-width action stack.
func (s *Server) MonsterActionBlockAttack532070(unit *Object, runtime MonsterActionBlockRuntime532070) bool {
	return monsterActionBlockAttack532070(unit, monsterActionBlockHooks532070{
		frame:      s.Frame,
		tickRate:   s.TickRate,
		testShield: runtime.TestShield,
		pop:        unit.MonsterPopAction,
		push:       unit.MonsterPushAction,
	})
}

func monsterActionBlockTerminal5320E0(unit *Object, pop func() int) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) || pop == nil {
		return false
	}
	if unit.UpdateDataMonster().Field120_3 != 0 {
		pop()
	}
	return true
}

// MonsterActionBlockFinish5320E0 restores GAME.EXE 005320E0.
func (s *Server) MonsterActionBlockFinish5320E0(unit *Object) bool {
	return monsterActionBlockTerminal5320E0(unit, unit.MonsterPopAction)
}

// MonsterActionWeaponBlock532110 restores GAME.EXE 00532110, whose behavior
// is identical to the shield-block finish action.
func (s *Server) MonsterActionWeaponBlock532110(unit *Object) bool {
	return monsterActionBlockTerminal5320E0(unit, unit.MonsterPopAction)
}
