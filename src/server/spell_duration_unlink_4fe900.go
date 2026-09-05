package server

// spellDurationUnlinkHooks4FE900 exposes every observable record-field load
// and list-link store in GAME.EXE 004FE900. Records are generic comparable
// tokens so the semantic core cannot inherit the original PE32 pointer width.
type spellDurationUnlinkHooks4FE900[Record comparable] struct {
	loadPrev  func(Record) Record
	loadNext  func(Record) Record
	storeNext func(Record, Record)
	storeHead func(Record)
	storePrev func(Record, Record)
}

// spellDurationUnlink4FE900 preserves GAME.EXE 004FE900's exact intrusive-
// list update order. Prev is read first. Next is deliberately reloaded after
// either the predecessor link or list head is written, and Prev is reloaded
// again only when that live Next is non-nil. No nil-record guard is added.
func spellDurationUnlink4FE900[Record comparable](
	record Record,
	h spellDurationUnlinkHooks4FE900[Record],
) {
	var nilRecord Record
	prev := h.loadPrev(record)
	if prev != nilRecord {
		next := h.loadNext(record)
		h.storeNext(prev, next)
	} else {
		next := h.loadNext(record)
		h.storeHead(next)
	}
	next := h.loadNext(record)
	if next != nilRecord {
		prev = h.loadPrev(record)
		h.storePrev(next, prev)
	}
}
