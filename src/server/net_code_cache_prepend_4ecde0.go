package server

// netCodeCachePrependHooks4ECDE0 separates the original list and entry
// pointer domains. Entry must use its zero value as the null link.
type netCodeCachePrependHooks4ECDE0[Entry comparable] struct {
	storeEntryPrev func(entry, prev Entry)
	loadFirst      func() Entry
	storeEntryNext func(entry, next Entry)
	storeLast      func(entry Entry)
	storeFirst     func(entry Entry)
}

// netCodeCachePrepend4ECDE0 preserves the load and store order of GAME.EXE
// 004ECDE0 without embedding an ABI32 pointer in an integer. In particular,
// the list head is loaded twice: the first value becomes entry.next, while
// the second value controls the head-prev versus list-last branch.
func netCodeCachePrepend4ECDE0[Entry comparable](entry Entry, hooks netCodeCachePrependHooks4ECDE0[Entry]) Entry {
	var nilEntry Entry
	hooks.storeEntryPrev(entry, nilEntry)
	first := hooks.loadFirst()
	hooks.storeEntryNext(entry, first)
	first = hooks.loadFirst()
	if first != nilEntry {
		hooks.storeEntryPrev(first, entry)
	} else {
		hooks.storeLast(entry)
	}
	hooks.storeFirst(entry)
	return entry
}
