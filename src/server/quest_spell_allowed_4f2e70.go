package server

const (
	questSpellRangeFirst4F2E70 = int32(75)
	questSpellRangeLast4F2E70  = int32(114)
)

type questSpellAllowedHooks4F2E70 struct {
	loadSpellID func(index int) uint32
	loadSlots   func(index int) uint32
}

func questSpellExplicitAllowed4F2E70(spellID int32) bool {
	switch spellID {
	case 46, 47, 48, 49, 122, 123, 124, 125:
		return true
	default:
		return false
	}
}

// questSpellAllowed4F2E70 reconstructs GAME.EXE 004F2E70. The original
// scans the live reward-spell table before applying its explicit-ID and
// signed-range fallbacks. Slots are read only for a matching SpellID; a match
// with zero Slots continues to the next row until the first zero-ID sentinel.
func questSpellAllowed4F2E70(
	spellID int32,
	hooks questSpellAllowedHooks4F2E70,
) int32 {
	index := 0
	currentSpellID := hooks.loadSpellID(index)
	tableAllowed := false
	if currentSpellID != 0 {
		target := uint32(spellID)
		for {
			if currentSpellID == target && hooks.loadSlots(index) != 0 {
				tableAllowed = true
				break
			}
			index++
			currentSpellID = hooks.loadSpellID(index)
			if currentSpellID == 0 {
				break
			}
		}
	}
	if questSpellExplicitAllowed4F2E70(spellID) {
		tableAllowed = true
	}
	if spellID >= questSpellRangeFirst4F2E70 && spellID <= questSpellRangeLast4F2E70 {
		return 1
	}
	if tableAllowed {
		return 1
	}
	return 0
}

// QuestSpellAllowed4F2E70 binds the predicate to the exact fixed-width
// reward-spell rows already reconstructed from GAME.EXE 005B9900.
//
//go:noinline
func QuestSpellAllowed4F2E70(spellID int32) int32 {
	return questSpellAllowed4F2E70(spellID, questSpellAllowedHooks4F2E70{
		loadSpellID: func(index int) uint32 {
			return rewardSpellDefinitions4F09F0[index].SpellID
		},
		loadSlots: func(index int) uint32 {
			return rewardSpellDefinitions4F09F0[index].Slots
		},
	})
}
