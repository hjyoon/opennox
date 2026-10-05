package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

// monsterMainHasShield5342C0 is the native-width form of GAME.EXE 005342C0.
// NPC equipment and ordinary monster status are alternative predicates, not
// interchangeable capabilities. The original service reads the live record.
func monsterMainHasShield5342C0(unit *Object) bool {
	if unit == nil || unit.UpdateData == nil {
		return false
	}
	update := unit.UpdateDataMonster()
	if unit.ObjSubClass.AsMonster().Has(object.MonsterNPC) {
		return update.ArmorEquipFlags&0x3000000 != 0
	}
	return update.StatusFlags.Has(object.MonStatusCanBlock)
}

// monsterMainBlock547210 restores GAME.EXE 005478AF..00547993. The entry
// stack-item pointer (replaced by the confusion dependency at 00547415) is
// retained across missile queries. Weapon flags use the cached update; the
// has-shield service and whole-stack BLOCK test use the live unit. Deadline
// clocks are read after normal pop/push callbacks. A rejected push still
// completes the original branch without writing a deadline.
func (s *Server) monsterMainBlock547210(unit *Object, update *MonsterUpdateData, head *AIStackItem, runtime MonsterMainRuntime547210) bool {
	if unit == nil || unit.UpdateData == nil || update == nil || head == nil || runtime.TestShield == nil {
		return false
	}
	if unit.ObjSubClass.AsMonster().Has(object.MonsterNPC) && update.WeaponEquipFlags&0x400 != 0 &&
		head.Type() != ai.ACTION_MELEE_ATTACK && head.Type() != ai.ACTION_MISSILE_ATTACK && runtime.TestShield(unit) != 0 {
		if head.Type() != ai.ACTION_WAIT && head.Type() != ai.ACTION_WEAPON_BLOCK {
			s.monsterMainPopAttackActions5471B0(unit)
			if wait := unit.MonsterPushAction(ai.ACTION_WAIT); wait != nil {
				wait.Args[0] = uintptr(s.Frame() + s.TickRate())
			}
		}
		return true
	}
	if monsterMainHasShield5342C0(unit) && head.Type() != ai.ACTION_MELEE_ATTACK &&
		head.Type() != ai.ACTION_MISSILE_ATTACK && runtime.TestShield(unit) != 0 {
		if !unit.UpdateDataMonster().HasAction(ai.ACTION_BLOCK_ATTACK) {
			s.monsterMainPopAttackActions5471B0(unit)
			if block := unit.MonsterPushAction(ai.ACTION_BLOCK_ATTACK); block != nil {
				block.Args[0] = uintptr(s.Frame() + (s.TickRate() >> 1))
			}
		}
		return true
	}
	return false
}
