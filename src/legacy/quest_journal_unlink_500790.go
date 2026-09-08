package legacy

// questJournalDeleteEntryHooks500790 exposes every observable journal-link
// load, list mutation, and release in GAME.EXE 00500790. Entry is a generic
// comparable token so the semantic contract cannot inherit PE32 pointer width.
type questJournalDeleteEntryHooks500790[Entry comparable] struct {
	loadPrev  func(Entry) Entry
	loadNext  func(Entry) Entry
	storeNext func(Entry, Entry)
	storePrev func(Entry, Entry)
	loadHead  func() Entry
	storeHead func(Entry)
	freeEntry func(Entry)
}

// questJournalDeleteEntryContract500790 preserves GAME.EXE 00500790's exact
// intrusive-list access order. Prev and Next are reloaded at every original
// memory read. The global head is changed only when its live identity equals
// entry, independently of whether entry.Prev is nil. The removed entry's own
// links remain untouched, release is last, and no nil-entry guard is added.
func questJournalDeleteEntryContract500790[Entry comparable](
	entry Entry,
	h questJournalDeleteEntryHooks500790[Entry],
) {
	var nilEntry Entry
	prev := h.loadPrev(entry)
	if prev != nilEntry {
		next := h.loadNext(entry)
		h.storeNext(prev, next)
	}
	next := h.loadNext(entry)
	if next != nilEntry {
		prev = h.loadPrev(entry)
		h.storePrev(next, prev)
	}
	if entry == h.loadHead() {
		next = h.loadNext(entry)
		h.storeHead(next)
	}
	h.freeEntry(entry)
}
