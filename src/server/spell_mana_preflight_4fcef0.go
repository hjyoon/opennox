package server

const (
	spellManaPreflightMonsterClass4FCEF0 = uint8(0x02)
	spellManaPreflightSummonFirst4FCEF0  = int32(75)
	spellManaPreflightSummonLast4FCEF0   = int32(114)
)

type spellManaPreflightHooks4FCEF0[Unit, Sequence comparable] struct {
	loadUnitArg     func() Unit
	loadSequenceArg func() Sequence
	loadCountArg    func() int32
	loadGodMode     func() int32
	loadClassLow    func(Unit) uint8
	loadOldMana     func(Unit) uint16
	loadSpell       func(Sequence, int32) int32
	summonCost      func(int32, Unit) int32
	spellManaCost   func(int32, int32) int32
}

// spellManaPreflight4FCEF0 preserves GAME.EXE 004FCEF0's gate, callback,
// and signed-arithmetic order without imposing PE32 pointer width on either
// pointer-bearing argument. A zero count fails before any engine or object
// observation, while a negative count deliberately reads current mana and
// then succeeds without touching the sequence. Positive counts have no
// built-in five-spell limit: callers own the readable sequence extent.
func spellManaPreflight4FCEF0[Unit, Sequence comparable](
	hooks spellManaPreflightHooks4FCEF0[Unit, Sequence],
) int32 {
	unit := hooks.loadUnitArg()
	var nilUnit Unit
	if unit == nilUnit {
		return 0
	}

	sequence := hooks.loadSequenceArg()
	var nilSequence Sequence
	if sequence == nilSequence {
		return 0
	}

	count := hooks.loadCountArg()
	if count == 0 {
		return 0
	}
	if hooks.loadGodMode() != 0 {
		return 1
	}
	if hooks.loadClassLow(unit)&spellManaPreflightMonsterClass4FCEF0 != 0 {
		return 1
	}

	remaining := int32(hooks.loadOldMana(unit))
	if count < 0 {
		return 1
	}

	for index := int32(0); ; index++ {
		spellID := hooks.loadSpell(sequence, index)
		var cost int32
		if spellID >= spellManaPreflightSummonFirst4FCEF0 && spellID <= spellManaPreflightSummonLast4FCEF0 {
			cost = hooks.summonCost(spellID, unit)
		} else {
			cost = hooks.spellManaCost(spellID, 2)
		}
		if cost > remaining {
			return 0
		}
		remaining = int32(uint32(remaining) - uint32(cost))
		if index+1 >= count {
			return 1
		}
	}
}
