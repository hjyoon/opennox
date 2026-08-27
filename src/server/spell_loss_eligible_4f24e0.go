package server

const (
	randomSpellLossExcludedGlyph4F24E0      = int32(34)
	randomSpellLossExcludedFireball4F24E0   = int32(27)
	randomSpellLossExcludedCharm4F24E0      = int32(9)
	randomSpellLossExcludedLesserHeal4F24E0 = int32(41)
)

type randomSpellLossEligibilityHooks4F24E0 struct {
	loadSpellID func(index int) uint32
	loadSlots   func(index int) uint32
}

// randomSpellLossEligible4F24E0 reconstructs GAME.EXE 004F24E0. The original
// starts at the SpellID member of the first twelve-byte reward-spell row. It
// scans live SpellID values until the zero-ID sentinel, reads Slots only for a
// matching ID, and accepts the first matching row whose Slots value is nonzero.
// Four learned starter spells remain protected from random loss even though
// their table rows have nonzero slot masks.
func randomSpellLossEligible4F24E0(
	spellID int32,
	hooks randomSpellLossEligibilityHooks4F24E0,
) int32 {
	index := 0
	currentSpellID := hooks.loadSpellID(index)
	if currentSpellID == 0 {
		return 0
	}
	target := uint32(spellID)
	for {
		if currentSpellID == target && hooks.loadSlots(index) != 0 {
			break
		}
		currentSpellID = hooks.loadSpellID(index + 1)
		index++
		if currentSpellID == 0 {
			return 0
		}
	}
	if spellID == 0 ||
		spellID == randomSpellLossExcludedGlyph4F24E0 ||
		spellID == randomSpellLossExcludedFireball4F24E0 ||
		spellID == randomSpellLossExcludedCharm4F24E0 ||
		spellID == randomSpellLossExcludedLesserHeal4F24E0 {
		return 0
	}
	return 1
}

// RandomSpellLossEligible4F24E0 binds the original predicate to the exact
// fixed-width reward-spell rows already reconstructed from GAME.EXE 005B9900.
//
//go:noinline
func RandomSpellLossEligible4F24E0(spellID int32) int32 {
	return randomSpellLossEligible4F24E0(spellID, randomSpellLossEligibilityHooks4F24E0{
		loadSpellID: func(index int) uint32 {
			return rewardSpellDefinitions4F09F0[index].SpellID
		},
		loadSlots: func(index int) uint32 {
			return rewardSpellDefinitions4F09F0[index].Slots
		},
	})
}
