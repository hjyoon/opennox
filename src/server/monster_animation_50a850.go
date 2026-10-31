package server

import (
	"github.com/opennox/libs/object"
	"unsafe"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func monsterAnimationIndex533790(update *MonsterUpdateData, mimic, plant, zombie bool) int {
	// Keep the older private helper's invalid-head containment. The raw
	// public entry retains the original signed native stack addressing.
	if update.AIStackInd < -1 || int(update.AIStackInd) >= len(update.AIStack) {
		update = &MonsterUpdateData{AIStackInd: -1, StatusFlags: update.StatusFlags, CurrentEnemy: update.CurrentEnemy}
	}
	predicate := func(value bool) func(*Object) uint32 {
		return func(*Object) uint32 {
			if value {
				return 1
			}
			return 0
		}
	}
	return int(MonsterAnimationNative533790(&Object{UpdateData: unsafe.Pointer(update)}, MonsterAnimationNativeRuntime533790{
		Mimic: predicate(mimic), Plant: predicate(plant), Zombie: predicate(zombie),
	}))
}

func monsterUpdateNonNPCAnim50A850(update *MonsterUpdateData, animInd int) {
	if update.Field120_3 != 0 || update.Field119 == nil || animInd < 0 || animInd >= len(update.Field119) {
		return
	}
	monsterUpdateAnim50A850(update, &update.Field119[animInd])
}

func monsterUpdateAnim50A850(update *MonsterUpdateData, anim *MonsterAnim) {
	update.Field120_0 = anim.frames
	if anim.frames == 0 {
		update.Field120_3 = 1
		return
	}
	update.Field120_2++
	if int(update.Field120_2) < int(anim.field10)+1 {
		return
	}
	update.Field120_2 = 0
	update.Field120_1++
	if update.Field120_1 < anim.frames {
		return
	}
	if anim.loop != 0 {
		update.Field120_1 = 0
		return
	}
	update.Field120_1 = anim.frames - 1
	update.Field120_3 = 1
}

func (s *Server) monsterNPCActionAnim533D00(update *MonsterUpdateData) MonsterAnim {
	var animInd int
	switch update.AIStackHead().Type() {
	case ai.ACTION_CAST_SPELL_ON_OBJECT, ai.ACTION_CAST_SPELL_ON_LOCATION, ai.ACTION_CAST_DURATION_SPELL:
		animInd = 21
	case ai.ACTION_DYING:
		animInd = 1
	case ai.ACTION_DEAD:
		animInd = 2
	case ai.ACTION_BLOCK_ATTACK:
		animInd = 40
	case ai.ACTION_WEAPON_BLOCK:
		if update.WeaponEquipFlags&0x7ff8000 != 0 {
			animInd = 30
		} else {
			animInd = 47
		}
	default:
		return MonsterAnim{}
	}
	frames, duration := s.PlayerAnimFrames(animInd)
	return MonsterAnim{frames: byte(frames), field10: byte(duration)}
}

// MonsterUpdateNPCAnim50A850 updates both native monster sprites and
// player-shaped NPCs without passing native pointers through the PE32 layout.
func (s *Server) MonsterUpdateNPCAnim50A850(unit *Object) bool {
	if unit == nil || unit.UpdateData == nil || !unit.ObjClass.Has(object.ClassMonster) {
		return false
	}
	update := unit.UpdateDataMonster()
	if unit.ObjSubClass.AsMonster().Has(object.MonsterNPC) {
		switch update.AIStackHead().Type() {
		case ai.ACTION_MELEE_ATTACK, ai.ACTION_MISSILE_ATTACK:
			// 0050A87E clears +483 even when a previous animation was
			// complete. Weapon attack owns the +481 frame and finish.
			update.Field120_3 = 0
			return true
		}
		if update.Field120_3 != 0 {
			return true
		}
		anim := s.monsterNPCActionAnim533D00(update)
		monsterUpdateAnim50A850(update, &anim)
		return true
	}
	ind := int(s.MonsterAnimationIndex533790(unit))
	monsterUpdateNonNPCAnim50A850(update, ind)
	return true
}
