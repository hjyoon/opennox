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

// MonsterHealSelection5411A0 carries the last eligible target through the C
// entry and standalone candidate callback.
type MonsterHealSelection5411A0 struct {
	Target *Object
}

// IsCasting5408A0 includes the original empty-stack check of Field132.
func (update *MonsterUpdateData) IsCasting5408A0() bool {
	action := ai.ActionType(update.Field132)
	if update.AIStackInd != -1 {
		action = ai.ActionType(update.AIStack[update.AIStackInd].Action)
	}
	return action >= ai.ACTION_CAST_SPELL_ON_OBJECT && action <= ai.ACTION_CAST_DURATION_SPELL
}

func monsterHealSomeone5411A0(unit *Object, hooks monsterHealHooks5411A0) int {
	var state MonsterHealSelection5411A0
	return state.healSomeone5411A0(unit, hooks)
}

func (state *MonsterHealSelection5411A0) healSomeone5411A0(unit *Object, hooks monsterHealHooks5411A0) int {
	if unit == nil || unit.UpdateData == nil || !unit.ObjFlags.Has(object.FlagEnabled) {
		return 0
	}
	update := unit.UpdateDataMonster()
	if !update.StatusFlags.Has(object.MonStatusCanCastSpells) || byte(hooks.frame())&0x1f != 0 {
		return 0
	}
	if update.IsCasting5408A0() {
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
	radius := float64(250)
	if hooks.quest() {
		radius = 640
	}
	state.Target = nil
	hooks.eachInRect(monsterHealSearchRect5411A0(unit.PosVec, radius), func(candidate *Object) bool {
		state.considerHealCandidate5412A0(candidate, unit, hooks.isEnemy, hooks.canInteract)
		return true
	})
	if state.Target == nil {
		return 0
	}
	hooks.cast(unit, monsterHealSpell5411A0, state.Target)
	return 1
}

func monsterHealSearchRect5411A0(pos types.Pointf, radius float64) types.Rectf {
	// The stock x87 rectangle spills each PC53 result to binary32 toward zero.
	edge := func(coordinate float32, delta float64) float32 {
		return float32(monsterMoveToRunSpill544440(monsterMoveToRunAddChop53_544434(float64(coordinate), delta)))
	}
	return types.Rectf{
		Min: types.Ptf(edge(pos.X, -radius), edge(pos.Y, -radius)),
		Max: types.Ptf(edge(pos.X, radius), edge(pos.Y, radius)),
	}
}

func (state *MonsterHealSelection5411A0) considerHealCandidate5412A0(candidate, caster *Object,
	isEnemy func(*Object, *Object) bool, canInteract func(*Object, *Object, int) bool) {
	if candidate == nil || candidate == caster || isEnemy(caster, candidate) ||
		candidate.HealthData == nil || candidate.ObjFlags.Has(object.FlagDead) || !canInteract(caster, candidate, 0) {
		return
	}
	health := candidate.HealthData
	if health.Cur < health.Max>>1 {
		state.Target = candidate
	}
}

// MonsterHealSomeone5411A0 restores the self/ally heal selection without
// passing native-width unit or health pointers through the PE32 C body.
func (s *Server) MonsterHealSomeone5411A0(unit *Object) int {
	var state MonsterHealSelection5411A0
	return s.MonsterHealWithSelection5411A0(unit, &state)
}

func (s *Server) MonsterHealWithSelection5411A0(unit *Object, state *MonsterHealSelection5411A0) int {
	return state.healSomeone5411A0(unit, monsterHealHooks5411A0{
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

func (s *Server) MonsterHealCandidate5412A0(candidate, caster *Object, state *MonsterHealSelection5411A0) {
	state.considerHealCandidate5412A0(candidate, caster, s.IsEnemyTo, s.CanInteract)
}
