package server

// spellDurationDuplicateHooks4FEE50 exposes every observable argument and
// duration-list load in GAME.EXE 004FEE50. Comparable record and object tokens
// preserve native null and pointer identity without inheriting the original
// PE32 pointer width.
type spellDurationDuplicateHooks4FEE50[Record, Object comparable] struct {
	loadHead         func() Record
	loadCasterArg    func() Object
	loadSpellArg     func() uint32
	loadFlag20       func(Record) uint32
	loadSpell        func(Record) uint32
	loadCaster       func(Record) Object
	loadFlagsLowByte func(Record) byte
	loadNext         func(Record) Record
}

// spellDurationDuplicate4FEE50 preserves GAME.EXE 004FEE50's exact
// short-circuit traversal. The list head is loaded first, and an empty list
// returns before either argument is read. On a non-empty list, caster is
// cached before the full spell dword. Each record then loads the full Flag20
// dword, full Spell dword, native caster identity, and low Flags88 byte only
// while all earlier predicates match. A match with flag bit zero clear returns
// canonical one without loading Next; every rejected record loads its live
// Next link last. Nil is a valid caster identity, and no record or cycle guard
// is added. Exhaustion returns canonical zero.
func spellDurationDuplicate4FEE50[Record, Object comparable](
	h spellDurationDuplicateHooks4FEE50[Record, Object],
) int32 {
	record := h.loadHead()
	var nilRecord Record
	if record == nilRecord {
		return 0
	}

	caster := h.loadCasterArg()
	requestedSpell := h.loadSpellArg()
	for {
		if h.loadFlag20(record) == 0 {
			if h.loadSpell(record) == requestedSpell {
				if h.loadCaster(record) == caster {
					if h.loadFlagsLowByte(record)&1 == 0 {
						return 1
					}
				}
			}
		}
		record = h.loadNext(record)
		if record == nilRecord {
			return 0
		}
	}
}
