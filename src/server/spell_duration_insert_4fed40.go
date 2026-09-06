package server

// spellDurationInsertHooks4FED40 exposes every observable argument load,
// duration-list head load, and intrusive-link store in GAME.EXE 004FED40.
// Comparable record tokens preserve native null values without imposing the
// original PE32 pointer width.
type spellDurationInsertHooks4FED40[Record comparable] struct {
	loadHead      func() Record
	loadRecordArg func() Record
	storePrev     func(Record, Record)
	storeNext     func(Record, Record)
	storeHead     func(Record)
}

// spellDurationInsert4FED40 preserves GAME.EXE 004FED40's exact head-
// insertion order. The old head is loaded before the record argument. When it
// is non-null, that cached old head receives the record as Prev. The record's
// Prev is then cleared, the list head is loaded again for the record's Next,
// and the record becomes the new head. No nil-record guard is added.
//
// The executable leaves the record in EAX, but its sole decoded caller at
// 004FECDB discards that residual value. This semantic boundary is therefore
// intentionally void.
func spellDurationInsert4FED40[Record comparable](
	h spellDurationInsertHooks4FED40[Record],
) {
	oldHead := h.loadHead()
	record := h.loadRecordArg()
	var nilRecord Record
	if oldHead != nilRecord {
		h.storePrev(oldHead, record)
	}
	h.storePrev(record, nilRecord)
	liveHead := h.loadHead()
	h.storeNext(record, liveHead)
	h.storeHead(record)
}
