package server

const (
	randomSpellCandidateCapacity4FE060 = 137
	randomSpellClassAny4FE060          = uint32(0x01000000)
)

type randomSpellSelectionHooks4FE060 struct {
	firstValid func() int32
	excluded   func(int32) int32
	flags      func(int32) uint32
	nextValid  func(int32) int32
	randomInt  func(int32, int32) int32
}

// randomSpellExcluded4FE100 reconstructs the exact 133-entry selector used by
// GAME.EXE 004FE100. The result is canonical zero or one for every signed
// dword input; values outside one through 133 take the original default path.
func randomSpellExcluded4FE100(spellID int32) int32 {
	switch spellID {
	case 1, 2, 6, 13, 15, 18, 19, 20, 30, 32, 33, 34, 38,
		51, 57, 68, 69, 70, 73, 129, 133:
		return 1
	default:
		return 0
	}
}

// RandomSpellExcluded4FE100 exposes the exact GAME.EXE 004FE100 exclusion
// predicate to the fixed-width legacy ABI without duplicating its selector.
func RandomSpellExcluded4FE100(spellID int32) int32 {
	return randomSpellExcluded4FE100(spellID)
}

// randomSpellSelection4FE060 reconstructs GAME.EXE 004FE060. Valid spells
// are traversed in registry order. The exclusion helper runs before the sole
// flags load, and NextValid runs for every nonzero spell regardless of which
// gate rejects it. A spell is retained only when its flags intersect the
// first mask or SpellClassAny and also intersect the second mask.
//
// The original fixed candidate array contains 137 signed dwords. Its RNG is
// called only for a nonempty result and is contractually inclusive, so its
// return is used directly as the candidate index.
func randomSpellSelection4FE060(
	firstMask uint32,
	secondMask uint32,
	hooks randomSpellSelectionHooks4FE060,
) int32 {
	var candidates [randomSpellCandidateCapacity4FE060]int32
	var count int32

	spellID := hooks.firstValid()
	if spellID == 0 {
		return 0
	}
	for {
		if hooks.excluded(spellID) == 0 {
			flags := hooks.flags(spellID)
			if (flags&firstMask != 0 || flags&randomSpellClassAny4FE060 != 0) &&
				flags&secondMask != 0 {
				candidates[count] = spellID
				count++
			}
		}
		spellID = hooks.nextValid(spellID)
		if spellID == 0 {
			break
		}
	}
	if count == 0 {
		return 0
	}
	return candidates[hooks.randomInt(0, count-1)]
}
