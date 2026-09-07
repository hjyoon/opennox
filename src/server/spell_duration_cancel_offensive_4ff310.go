package server

const spellDurationOffensiveLowByte4FF310 = byte(0x20)

// SpellDurationCancelOffensiveHooks4FF310 exposes every observable list,
// argument, record, and callback access in GAME.EXE 004FF310. Comparable
// record and object tokens preserve canonical null values and full native
// pointer identities without inheriting PE32's pointer width.
type SpellDurationCancelOffensiveHooks4FF310[Record, Object comparable] struct {
	LoadFirst      func() Record
	LoadCasterArg  func() Object
	LoadCaster     func(Record) Object
	LoadNext       func(Record) Record
	LoadSpell      func(Record) int32
	LoadSpellFlags func(int32) uint32
	Cancel         func(Record)
}

// SpellDurationCancelOffensive4FF310 preserves GAME.EXE 004FF310's exact
// traversal and short-circuit order. The list head is loaded before the
// caster argument, so an empty list returns without reading that argument.
// On a non-empty list the requested caster identity is cached once. For each
// record, Caster16 is loaded before Next, and Next is snapshotted before any
// comparison or callback. A caster mismatch skips both Spell and callbacks.
// A match passes the complete signed dword spell ID to the flags callback and
// cancels the current record only when AL has bit 0x20 set; upper return bytes
// are ignored. Traversal always follows the saved successor even if either
// callback mutates the current record or list. Nil is a valid caster identity,
// and no record, callback, or cycle guard is added.
func SpellDurationCancelOffensive4FF310[Record, Object comparable](
	h SpellDurationCancelOffensiveHooks4FF310[Record, Object],
) {
	record := h.LoadFirst()
	var nilRecord Record
	if record == nilRecord {
		return
	}

	casterArg := h.LoadCasterArg()
	for {
		caster := h.LoadCaster(record)
		next := h.LoadNext(record)
		if caster == casterArg {
			spellID := h.LoadSpell(record)
			if byte(h.LoadSpellFlags(spellID))&spellDurationOffensiveLowByte4FF310 != 0 {
				h.Cancel(record)
			}
		}
		record = next
		if record == nilRecord {
			return
		}
	}
}
