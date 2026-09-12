package server

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func monsterScriptHitNativeHooks515A30(canAttack func(*Object) bool) monsterScriptHitHooks515A30[*Object, *AIStackItem] {
	return monsterScriptHitHooks515A30[*Object, *AIStackItem]{
		loadClassLow: func(unit *Object) uint8 { return uint8(unit.ObjClass) },
		loadFlags:    func(unit *Object) uint32 { return uint32(unit.ObjFlags) },
		canAttack:    canAttack,
		clearActionStack: func(unit *Object) {
			unit.ClearActionStack()
		},
		pushAction: func(unit *Object, action uint32) *AIStackItem {
			return unit.MonsterPushAction(ai.ActionType(action))
		},
		storeArgBits: func(item *AIStackItem, index int, bits uint32) {
			item.Args[index] = uintptr(bits)
		},
		loadMeleeRange: func(unit *Object) float32 {
			return unit.UpdateDataMonster().MonsterDef.MeleeAttackRange112
		},
		loadRadius: func(unit *Object) float32 {
			return unit.Shape.Circle.R
		},
	}
}

func monsterScriptHitMeleeNative515A30(unit *Object, pos *types.Pointf) {
	monsterScriptHitMelee515A30(unit, pos, monsterScriptHitNativeHooks515A30(func(unit *Object) bool {
		return monsterFightCanMelee534220(unit, unit.UpdateDataMonster())
	}))
}

func monsterScriptHitMissileNative515B80(unit *Object, pos *types.Pointf) {
	monsterScriptHitMissile515B80(unit, pos, monsterScriptHitNativeHooks515A30(func(unit *Object) bool {
		return monsterFightCanShoot534280(unit, unit.UpdateDataMonster())
	}))
}

// MonsterScriptHitMelee515A30 replaces the legacy PE32 object-int conversion
// while retaining the original script-facing action sequence.
func (*Server) MonsterScriptHitMelee515A30(unit *Object, pos types.Pointf) {
	monsterScriptHitMeleeNative515A30(unit, &pos)
}

// MonsterScriptHitMissile515B80 replaces the crashing PE32 object-int
// conversion; its position is an ABI32 float pair, never an object address.
func (*Server) MonsterScriptHitMissile515B80(unit *Object, pos types.Pointf) {
	monsterScriptHitMissileNative515B80(unit, &pos)
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.ObjClass)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.ObjFlags)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(MonsterDef{}.MeleeAttackRange112)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.Shape.Circle.R)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(AIStackItem{}.Action)]
)
