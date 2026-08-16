package server

const netCodeCacheInitArrayCapacity4ECE50 = 16

// netCodeCacheInitArrayHooks4ECE50 separates the original cache-entry pointer
// domain from the four list-head fields and the initialization flag. Entry and
// Result may remain native pointers; neither is narrowed to an ABI32 integer.
type netCodeCacheInitArrayHooks4ECE50[Entry any, Result any] struct {
	storeUsedFirst func(Entry)
	storeUsedLast  func(Entry)
	storeFreeFirst func(Entry)
	storeFreeLast  func(Entry)
	prependFree    func(Entry) Result
	clearNeedsInit func()
}

// netCodeCacheInitArray4ECE50 preserves the store, call, and return order of
// GAME.EXE 004ECE50. It deliberately does not clear entry values: the original
// only resets both list pairs, prepends all sixteen entries in address order,
// and clears the needs-initialization flag after the final prepend succeeds.
func netCodeCacheInitArray4ECE50[Entry any, Result any](entries [netCodeCacheInitArrayCapacity4ECE50]Entry, hooks netCodeCacheInitArrayHooks4ECE50[Entry, Result]) Result {
	var nilEntry Entry
	hooks.storeUsedFirst(nilEntry)
	hooks.storeUsedLast(nilEntry)
	hooks.storeFreeFirst(nilEntry)
	hooks.storeFreeLast(nilEntry)

	var result Result
	for i := 0; i < len(entries); i++ {
		result = hooks.prependFree(entries[i])
	}
	hooks.clearNeedsInit()
	return result
}
