package server

// PlayerCancelSpellsHooks4FEAE0 exposes every observable list access and
// callback in GAME.EXE 004FEAE0. Comparable record and object tokens preserve
// native null and pointer identity without inheriting the original PE32 width.
type PlayerCancelSpellsHooks4FEAE0[Record, Object comparable] struct {
	LoadFirst  func() Record
	LoadCaster func(Record) Object
	LoadNext   func(Record) Record
	Cancel     func(Record)
}

// PlayerCancelSpells4FEAE0 preserves GAME.EXE 004FEAE0's traversal order.
// Each record's caster is loaded before its successor is snapshotted. A
// matching record is cancelled only after that snapshot, so cancellation
// cannot redirect the current traversal. Nil is a valid caster identity, and
// no record, callback, or cycle guard is added. Both the empty and exhausted
// paths return the original canonical zero.
func PlayerCancelSpells4FEAE0[Record, Object comparable](
	caster Object,
	h PlayerCancelSpellsHooks4FEAE0[Record, Object],
) int32 {
	record := h.LoadFirst()
	var nilRecord Record
	if record == nilRecord {
		return 0
	}

	for {
		recordCaster := h.LoadCaster(record)
		next := h.LoadNext(record)
		if recordCaster == caster {
			h.Cancel(record)
		}
		record = next
		if record == nilRecord {
			return 0
		}
	}
}
