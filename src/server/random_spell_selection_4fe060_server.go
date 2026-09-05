package server

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
)

type randomSpellSelectionNativeDeps4FE060 struct {
	firstValid func() spell.ID
	flags      func(spell.ID) things.SpellFlags
	nextValid  func(spell.ID) spell.ID
	randomInt  func(int32, int32) int32
}

func randomSpellSelectionNative4FE060(
	firstMask uint32,
	secondMask uint32,
	deps randomSpellSelectionNativeDeps4FE060,
) int32 {
	return randomSpellSelection4FE060(firstMask, secondMask, randomSpellSelectionHooks4FE060{
		firstValid: func() int32 {
			return int32(deps.firstValid())
		},
		excluded: randomSpellExcluded4FE100,
		flags: func(spellID int32) uint32 {
			return uint32(deps.flags(spell.ID(spellID)))
		},
		nextValid: func(spellID int32) int32 {
			return int32(deps.nextValid(spell.ID(spellID)))
		},
		randomInt: deps.randomInt,
	})
}

// RandomSpell4FE060 binds GAME.EXE 004FE060 to the native spell registry and
// the server logic RNG. Spell IDs cross the restored ABI as signed dwords and
// spell flags and both masks remain full unsigned dwords.
//
//go:noinline
func (s *Server) RandomSpell4FE060(firstMask, secondMask uint32) int32 {
	return randomSpellSelectionNative4FE060(firstMask, secondMask, randomSpellSelectionNativeDeps4FE060{
		firstValid: s.Spells.FirstValid,
		flags:      s.Spells.Flags,
		nextValid:  s.Spells.NextValid,
		randomInt: func(minimum, maximum int32) int32 {
			return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
		},
	})
}
