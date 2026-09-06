package server

// SpellCancelDurSpellHooks4FEB10 exposes every observable list access and
// callback in GAME.EXE 004FEB10. Comparable record and object tokens preserve
// native null and pointer identity without inheriting the original PE32 width.
type SpellCancelDurSpellHooks4FEB10[Record, Object comparable] struct {
	LoadFirst  func() Record
	LoadSpell  func(Record) int32
	LoadNext   func(Record) Record
	LoadCaster func(Record) Object
	Cancel     func(Record)
}

// SpellCancelDurSpell4FEB10 preserves GAME.EXE 004FEB10's signed-ID tests and
// traversal order. Each record's spell is loaded before its successor is
// snapshotted. An exact spell match tests the live caster first. If that test
// fails and both IDs are in the inclusive summon range 75 through 114, the
// live caster is loaded a second time before deciding whether to cancel. A
// callback runs only after the successor snapshot, so it cannot redirect the
// current traversal. Nil is a valid caster identity, and no record, callback,
// or cycle guard is added. Both the empty and exhausted paths return the
// original canonical zero.
func SpellCancelDurSpell4FEB10[Record, Object comparable](
	requestedSpell int32,
	caster Object,
	h SpellCancelDurSpellHooks4FEB10[Record, Object],
) int32 {
	record := h.LoadFirst()
	var nilRecord Record
	if record == nilRecord {
		return 0
	}

	for {
		recordSpell := h.LoadSpell(record)
		next := h.LoadNext(record)
		cancel := false
		if recordSpell == requestedSpell {
			cancel = h.LoadCaster(record) == caster
		}
		if !cancel &&
			requestedSpell >= 75 && requestedSpell <= 114 &&
			recordSpell >= 75 && recordSpell <= 114 {
			cancel = h.LoadCaster(record) == caster
		}
		if cancel {
			h.Cancel(record)
		}
		record = next
		if record == nilRecord {
			return 0
		}
	}
}
