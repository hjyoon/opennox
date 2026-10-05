package server

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

const monsterFleeRelatedSpellMask541050 = uint32(0x80000000)

// monsterFleeCastRelated541050 preserves the fourth spell selector, 00541050.
// Unlike 00540D90 it targets self, suppresses an already active enchant, and
// does not check summon limits. Cooldown RNG precedes the post-cast frame read.
func monsterFleeCastRelated541050(unit *Object, hooks monsterFightSpellHooks540B90) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) ||
		hooks.frame == nil || hooks.spellAllowed == nil || hooks.random == nil || hooks.cast == nil {
		return false
	}
	update := unit.UpdateDataMonster()
	if !unit.Flags().Has(object.FlagEnabled) || !unit.UpdateDataMonster().StatusFlags.Has(object.MonStatusCanCastSpells) ||
		hooks.frame() < update.Field371 {
		return false
	}
	head := unit.UpdateDataMonster().AIStackHead()
	if head == nil || (head.Type() >= ai.ACTION_CAST_SPELL_ON_OBJECT && head.Type() <= ai.ACTION_CAST_DURATION_SPELL) {
		return false
	}
	candidates := monsterFightSpellCandidates540B90(update, monsterFleeRelatedSpellMask541050, hooks.spellAllowed)
	if len(candidates) == 0 {
		return false
	}
	for _, id := range candidates {
		if monsterFightHasSpellEnchant540CE0(unit, id) {
			return false
		}
	}
	id := candidates[hooks.random(0, len(candidates)-1)]
	hooks.cast(unit, id, unit)
	cooldown := hooks.random(int(update.Field370_0), int(update.Field370_2))
	update.Field371 = hooks.frame() + uint32(cooldown)
	return true
}
