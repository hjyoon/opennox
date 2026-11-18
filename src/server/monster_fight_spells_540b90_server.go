package server

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
)

func (s *Server) monsterFightSpellHooks540B90(canSummon func(*Object, int) bool) monsterFightSpellHooks540B90 {
	return monsterFightSpellHooks540B90{
		frame: s.Frame,
		spellAllowed: func(id spell.ID) bool {
			return s.Spells.HasFlags(id, things.SpellMobsCanCast)
		},
		random: s.Rand.Logic.IntClamp,
		cast: func(unit *Object, id spell.ID, target *Object) {
			unit.MonsterCast(id, target)
		},
		activeSummon: s.Spells.Dur.HasCasterSummon540D40,
		canSummon:    canSummon,
	}
}

func (sp *SpellsDuration) HasCasterSummon540D40(unit *Object) bool {
	for record := sp.SpellDurationFirst4FE930(); record != nil; record = SpellDurationNextNative4FE940(record) {
		if MonsterSpellIsSummon540D20(spell.ID(record.Spell)) && record.Caster16 == unit {
			return true
		}
	}
	return false
}
