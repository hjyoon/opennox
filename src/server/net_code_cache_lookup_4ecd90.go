package server

// netCodeCacheLookupHooks4ECD90 separates the original cache-entry pointer
// domain from the cached object pointer domain. Entry must use its zero value
// as the null link, matching the original linked-list representation.
type netCodeCacheLookupHooks4ECD90[Entry comparable, Object any] struct {
	loadNeedsInit func() bool
	initCache     func()
	loadFirstUsed func() Entry
	loadObject    func(Entry) Object
	loadNetCode   func(Object) uint32
	loadNext      func(Entry) Entry
	removeUsed    func(Entry)
	prependUsed   func(Entry)
}

// netCodeCacheLookupObject4ECD90 preserves the load and mutation order of
// GAME.EXE 004ECD90 without embedding an ABI32 pointer in an integer.
func netCodeCacheLookupObject4ECD90[Entry comparable, Object any](code uint32, hooks netCodeCacheLookupHooks4ECD90[Entry, Object]) Object {
	if hooks.loadNeedsInit() {
		hooks.initCache()
	}
	entry := hooks.loadFirstUsed()
	var nilEntry Entry
	var nilObject Object
	if entry == nilEntry {
		return nilObject
	}
	for {
		obj := hooks.loadObject(entry)
		if hooks.loadNetCode(obj) == code {
			hooks.removeUsed(entry)
			hooks.prependUsed(entry)
			return hooks.loadObject(entry)
		}
		entry = hooks.loadNext(entry)
		if entry == nilEntry {
			return nilObject
		}
	}
}
