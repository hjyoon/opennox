package server

// netCodeCacheNextUnusedHooks4ECEF0 separates the free-list head from cache
// entry identities. Entry may remain a native pointer on every target.
type netCodeCacheNextUnusedHooks4ECEF0[Entry comparable] struct {
	loadFirstFree  func() Entry
	loadEntryNext  func(Entry) Entry
	storeFirstFree func(Entry)
}

// netCodeCacheNextUnused4ECEF0 preserves GAME.EXE 004ECEF0. It loads the free
// head once and returns nil immediately when empty. Otherwise it loads next
// from that cached entry, publishes next as the new head, and returns the
// original entry without changing the free tail or either entry link.
func netCodeCacheNextUnused4ECEF0[Entry comparable](hooks netCodeCacheNextUnusedHooks4ECEF0[Entry]) Entry {
	entry := hooks.loadFirstFree()
	var nilEntry Entry
	if entry == nilEntry {
		return nilEntry
	}
	next := hooks.loadEntryNext(entry)
	hooks.storeFirstFree(next)
	return entry
}
