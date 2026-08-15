package server

// netCodeCacheRemoveHooks4ECE10 separates the original list and entry
// pointer domains. Entry must use its zero value as the null link.
type netCodeCacheRemoveHooks4ECE10[Entry comparable] struct {
	loadEntryNext  func(entry Entry) Entry
	loadEntryPrev  func(entry Entry) Entry
	storeEntryPrev func(entry, prev Entry)
	storeLast      func(entry Entry)
	storeEntryNext func(entry, next Entry)
	storeFirst     func(entry Entry)
}

// netCodeCacheRemove4ECE10 preserves the load, store, and return order of
// GAME.EXE 004ECE10 without embedding an ABI32 pointer in an integer. The
// entry links are deliberately loaded more than once: the first next and
// first prev values repair the successor or list tail, while independently
// reloaded prev and next values repair the predecessor or list head.
func netCodeCacheRemove4ECE10[Entry comparable](entry Entry, hooks netCodeCacheRemoveHooks4ECE10[Entry]) Entry {
	var nilEntry Entry
	next := hooks.loadEntryNext(entry)
	if next != nilEntry {
		prev := hooks.loadEntryPrev(entry)
		hooks.storeEntryPrev(next, prev)
	} else {
		prev := hooks.loadEntryPrev(entry)
		hooks.storeLast(prev)
	}
	prev := hooks.loadEntryPrev(entry)
	if prev != nilEntry {
		next = hooks.loadEntryNext(entry)
		hooks.storeEntryNext(prev, next)
		return entry
	}
	next = hooks.loadEntryNext(entry)
	hooks.storeFirst(next)
	return next
}
