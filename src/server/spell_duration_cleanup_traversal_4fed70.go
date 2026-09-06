package server

// SpellDurationCleanupTraversalHooks4FED70 exposes every observable list
// access and callback in GAME.EXE 004FED70. Comparable record tokens preserve
// native null and pointer identity without inheriting the original PE32 width.
type SpellDurationCleanupTraversalHooks4FED70[Record comparable] struct {
	LoadFirst        func() Record
	LoadFlagsLowByte func(Record) byte
	LoadNext         func(Record) Record
	Destroy          func(Record)
}

// SpellDurationCleanupTraversal4FED70 preserves GAME.EXE 004FED70's exact
// traversal order. Each record's low flag byte is loaded before its successor
// is snapshotted. A record whose saved flag byte has bit zero set is destroyed
// only after that snapshot, so destruction cannot redirect the traversal. No
// record, callback, or cycle guard is added. The executable's residual EAX is
// not part of the void caller contract.
func SpellDurationCleanupTraversal4FED70[Record comparable](
	h SpellDurationCleanupTraversalHooks4FED70[Record],
) {
	record := h.LoadFirst()
	var nilRecord Record
	if record == nilRecord {
		return
	}

	for {
		flags := h.LoadFlagsLowByte(record)
		next := h.LoadNext(record)
		if flags&1 != 0 {
			h.Destroy(record)
		}
		record = next
		if record == nilRecord {
			return
		}
	}
}
