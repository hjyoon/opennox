package server

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

type monsterFleeSpellHooks544760 struct {
	buffSelf      func(*Object) bool
	castRelated   func(*Object) bool
	canInteract   func(*Object, *Object, int) bool
	castOffensive func(*Object, *Object) bool
}

// monsterFleeTrySpell544760 is the enemy-present spell branch of 00544760.
// Eligibility uses live update data, but enemy reloads use the caller's cached
// entry update. Success suppresses offensive selection, not subsequent movement.
func monsterFleeTrySpell544760(unit *Object, cached *MonsterUpdateData, hooks monsterFleeSpellHooks544760) {
	if unit == nil || unit.UpdateData == nil || cached == nil {
		return
	}
	live := unit.UpdateDataMonster()
	if !live.StatusFlags.Has(object.MonStatusCanCastSpells) || unit.HasEnchant(ENCHANT_ANTI_MAGIC) {
		return
	}
	aggression := live.Aggression
	if (aggression < monsterMainActiveAggressionLimit547210 && aggression > monsterMainPassiveAggressionLimit547210) ||
		aggression < monsterMainPassiveAggressionLimit547210 {
		return
	}
	if live.HasAction(ai.DEPENDENCY_UNINTERRUPTABLE) {
		if hooks.castRelated == nil || hooks.castRelated(unit) {
			return
		}
	} else if hooks.buffSelf == nil || hooks.buffSelf(unit) {
		return
	}
	if hooks.canInteract != nil && hooks.canInteract(unit, cached.CurrentEnemy, 0) && hooks.castOffensive != nil {
		hooks.castOffensive(unit, cached.CurrentEnemy)
	}
}

func (s *Server) monsterFleeTrySpell544760(unit *Object, cached *MonsterUpdateData) {
	monsterFleeTrySpell544760(unit, cached, monsterFleeSpellHooks544760{
		buffSelf: func(unit *Object) bool {
			return monsterFightBuffSelf540B90(unit, s.monsterFightSpellHooks540B90(nil))
		},
		castRelated: func(unit *Object) bool {
			return monsterFleeCastRelated541050(unit, s.monsterFightSpellHooks540B90(nil))
		},
		canInteract: s.CanInteract,
		castOffensive: func(unit, target *Object) bool {
			return monsterFightCastOffensive540F20(unit, target, s.monsterFightSpellHooks540B90(nil))
		},
	})
}
