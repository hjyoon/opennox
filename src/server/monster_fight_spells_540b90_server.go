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

func (s *Server) MonsterFightBuffSelf540B90(unit *Object) bool {
	return monsterFightBuffSelf540B90(unit, s.monsterFightSpellHooks540B90(nil))
}

func (s *Server) MonsterFightCastOffensive540F20(unit, target *Object) bool {
	return monsterFightCastOffensive540F20(unit, target, s.monsterFightSpellHooks540B90(nil))
}

func (s *Server) MonsterFightCastRelated540D90(unit, target *Object, canSummon func(*Object, int) bool) bool {
	return monsterFightCastRelated540D90(unit, target, s.monsterFightSpellHooks540B90(canSummon))
}

func (s *Server) MonsterFleeCastRelated541050(unit *Object) bool {
	return monsterFleeCastRelated541050(unit, s.monsterFightSpellHooks540B90(nil))
}

func (sp *SpellsDuration) HasCasterSummon540D40(unit *Object) bool {
	for record := sp.SpellDurationFirst4FE930(); record != nil; record = SpellDurationNextNative4FE940(record) {
		if MonsterSpellIsSummon540D20(spell.ID(record.Spell)) && record.Caster16 == unit {
			return true
		}
	}
	return false
}
