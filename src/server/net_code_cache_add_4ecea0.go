package server

// netCodeCacheAddHooks4ECEA0 separates cache-entry and object identities from
// the original ABI32 list operations. Entry may remain a native pointer and
// Object is never converted through an integer-width boundary.
type netCodeCacheAddHooks4ECEA0[Entry comparable, Object, Result any] struct {
	nextUnused   func() Entry
	loadObject   func() Object
	loadLastUsed func() Entry
	storeObject  func(Entry, Object)
	removeUsed   func(Entry)
	prependUsed  func(Entry) Result
}

// netCodeCacheAddObject4ECEA0 preserves the branch, load, store, call, and
// return order of GAME.EXE 004ECEA0. The object argument is observed only
// after nextUnused returns. When the free list is empty, the used-list tail is
// loaded once, overwritten before removal, and the same cached entry is then
// prepended. The prepend result is returned on both paths.
func netCodeCacheAddObject4ECEA0[Entry comparable, Object, Result any](hooks netCodeCacheAddHooks4ECEA0[Entry, Object, Result]) Result {
	entry := hooks.nextUnused()
	obj := hooks.loadObject()

	var nilEntry Entry
	if entry == nilEntry {
		entry = hooks.loadLastUsed()
		hooks.storeObject(entry, obj)
		hooks.removeUsed(entry)
		return hooks.prependUsed(entry)
	}

	hooks.storeObject(entry, obj)
	return hooks.prependUsed(entry)
}
