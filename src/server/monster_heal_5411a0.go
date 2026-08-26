package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

const monsterHealSpell5411A0 = spell.ID(41)

type monsterHealHooks5411A0 struct {
	frame       func() uint32
	quest       func() bool
	eachInRect  func(types.Rectf, func(*Object) bool)
	isEnemy     func(*Object, *Object) bool
	canInteract func(*Object, *Object, int) bool
	cast        func(*Object, spell.ID, *Object)
}

func monsterHealSomeone5411A0(unit *Object, hooks monsterHealHooks5411A0) int {
	if unit == nil || unit.UpdateData == nil || !unit.ObjFlags.Has(object.FlagEnabled) {
		return 0
	}
	update := unit.UpdateDataMonster()
	if !update.StatusFlags.Has(object.MonStatusCanCastSpells) || byte(hooks.frame())&0x1f != 0 {
		return 0
	}
	if action := update.AIStackHead().Type(); action >= ai.ACTION_CAST_SPELL_ON_OBJECT && action <= ai.ACTION_CAST_DURATION_SPELL {
		return 0
	}
	if update.StatusFlags.Has(object.MonStatusCanHealSelf) {
		if health := unit.HealthData; health != nil && health.Max != 0 && health.Cur < health.Max>>1 {
			hooks.cast(unit, monsterHealSpell5411A0, unit)
			return 0
		}
	}
	if !update.StatusFlags.Has(object.MonStatusCanHealOthers) {
		return 0
	}
	radius := float32(250)
	if hooks.quest() {
		radius = 640
	}
	pos := unit.PosVec
	var target *Object
	hooks.eachInRect(types.Rectf{
		Min: types.Ptf(pos.X-radius, pos.Y-radius),
		Max: types.Ptf(pos.X+radius, pos.Y+radius),
	}, func(candidate *Object) bool {
		if candidate == nil || candidate == unit || hooks.isEnemy(unit, candidate) ||
			candidate.HealthData == nil || candidate.ObjFlags.Has(object.FlagDead) ||
			!hooks.canInteract(unit, candidate, 0) {
			return true
		}
		health := candidate.HealthData
		if health.Cur < health.Max>>1 {
			target = candidate
		}
		return true
	})
	if target == nil {
		return 0
	}
	hooks.cast(unit, monsterHealSpell5411A0, target)
	return 1
}

// MonsterHealSomeone5411A0 restores the self/ally heal selection without
// passing native-width unit or health pointers through the PE32 C body.
func (s *Server) MonsterHealSomeone5411A0(unit *Object) int {
	return monsterHealSomeone5411A0(unit, monsterHealHooks5411A0{
		frame: s.Frame,
		quest: func() bool {
			return noxflags.HasGame(noxflags.GameModeQuest)
		},
		eachInRect:  s.Map.EachObjInRect,
		isEnemy:     s.IsEnemyTo,
		canInteract: s.CanInteract,
		cast: func(caster *Object, id spell.ID, target *Object) {
			caster.MonsterCast(id, target)
		},
	})
}
