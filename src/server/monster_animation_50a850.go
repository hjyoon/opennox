package server

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func monsterAnimationIndex533790(update *MonsterUpdateData, mimic, plant, zombie bool) int {
	ind := 8
	if head := update.AIStackHead(); head != nil {
		switch head.Type() {
		case ai.ACTION_MOVE_TO, ai.ACTION_FAR_MOVE_TO, ai.ACTION_ROAM, ai.ACTION_FIND_OBJECT,
			ai.ACTION_RANDOM_WALK, ai.ACTION_CONFUSED, ai.ACTION_MOVE_TO_HOME:
			ind = 12
			if update.StatusFlags.Has(object.MonStatusRunning) {
				ind = 13
			}
		case ai.ACTION_DODGE:
			ind = 12
		case ai.ACTION_MELEE_ATTACK:
			ind = 1
		case ai.ACTION_MISSILE_ATTACK:
			ind = 3
		case ai.ACTION_CAST_SPELL_ON_OBJECT, ai.ACTION_CAST_SPELL_ON_LOCATION, ai.ACTION_CAST_DURATION_SPELL:
			ind = 7
		case ai.ACTION_BLOCK_ATTACK, ai.ACTION_WEAPON_BLOCK:
			ind = 5
		case ai.ACTION_BLOCK_FINISH:
			ind = 6
		case ai.ACTION_FLEE:
			ind = 13
		case ai.ACTION_DYING:
			ind = 9
		case ai.ACTION_DEAD:
			ind = 10
		case ai.ACTION_MORPH_INTO_CHEST, ai.ACTION_GET_UP:
			ind = 14
		case ai.ACTION_MORPH_BACK_TO_SELF:
			ind = 15
		}
	}
	if mimic && ind == 8 && update.StatusFlags.Has(object.MonStatusMorphed) {
		return 0
	}
	if plant && ind == 8 && update.CurrentEnemy == nil {
		return 14
	}
	if zombie && ind == 9 && update.StatusFlags.Has(object.MonStatusOnFire) {
		return 15
	}
	return ind
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
		if update.Field120_3 != 0 {
			return true
		}
		switch update.AIStackHead().Type() {
		case ai.ACTION_MELEE_ATTACK, ai.ACTION_MISSILE_ATTACK:
			update.Field120_3 = 1
			return true
		}
		anim := s.monsterNPCActionAnim533D00(update)
		monsterUpdateAnim50A850(update, &anim)
		return true
	}
	ind := monsterAnimationIndex533790(update, s.IsMimic(unit), s.IsPlant(unit), s.IsZombie(unit))
	monsterUpdateNonNPCAnim50A850(update, ind)
	return true
}
