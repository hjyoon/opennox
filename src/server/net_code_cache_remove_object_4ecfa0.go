package server

type netCodeCacheRemoveObjectResultKind4ECFA0 uint8

const (
	netCodeCacheRemoveObjectInitial4ECFA0 netCodeCacheRemoveObjectResultKind4ECFA0 = iota
	netCodeCacheRemoveObjectArgument4ECFA0
	netCodeCacheRemoveObjectEntry4ECFA0
)

// netCodeCacheRemoveObjectResult4ECFA0 keeps the original mixed EAX return
// domains distinct. GAME.EXE returns the raw initialization flag before the
// cache can be searched, zero for an initialized empty list, the object
// argument after a miss, and the free-list prepend result after a hit.
type netCodeCacheRemoveObjectResult4ECFA0[Entry, Object any] struct {
	kind        netCodeCacheRemoveObjectResultKind4ECFA0
	initialFlag uint32
	object      Object
	entry       Entry
}

// netCodeCacheRemoveObjectHooks4ECFA0 separates cache-entry and object
// identities from the original ABI32 list storage. Both identities can stay
// native pointers on every target architecture.
type netCodeCacheRemoveObjectHooks4ECFA0[Entry, Object comparable] struct {
	loadNeedsInit   func() uint32
	loadFirstUsed   func() Entry
	loadObjectArg   func() Object
	loadEntryObject func(Entry) Object
	loadEntryNext   func(Entry) Entry
	removeUsed      func(Entry)
	prependFree     func(Entry) Entry
}

// netCodeCacheRemoveObject4ECFA0 preserves GAME.EXE 004ECFA0. The
// initialization flag is loaded before any list or argument access. Once the
// used-list head is known to be non-null, the object argument is loaded once
// and entries are compared before their next links. A hit removes the exact
// cached entry and prepends it to the free list without clearing its object.
func netCodeCacheRemoveObject4ECFA0[Entry, Object comparable](hooks netCodeCacheRemoveObjectHooks4ECFA0[Entry, Object]) netCodeCacheRemoveObjectResult4ECFA0[Entry, Object] {
	needsInit := hooks.loadNeedsInit()
	if needsInit != 0 {
		return netCodeCacheRemoveObjectResult4ECFA0[Entry, Object]{
			kind:        netCodeCacheRemoveObjectInitial4ECFA0,
			initialFlag: needsInit,
		}
	}

	entry := hooks.loadFirstUsed()
	var nilEntry Entry
	if entry == nilEntry {
		return netCodeCacheRemoveObjectResult4ECFA0[Entry, Object]{
			kind:        netCodeCacheRemoveObjectInitial4ECFA0,
			initialFlag: needsInit,
		}
	}

	obj := hooks.loadObjectArg()
	for {
		if hooks.loadEntryObject(entry) == obj {
			hooks.removeUsed(entry)
			return netCodeCacheRemoveObjectResult4ECFA0[Entry, Object]{
				kind:  netCodeCacheRemoveObjectEntry4ECFA0,
				entry: hooks.prependFree(entry),
			}
		}
		entry = hooks.loadEntryNext(entry)
		if entry == nilEntry {
			return netCodeCacheRemoveObjectResult4ECFA0[Entry, Object]{
				kind:   netCodeCacheRemoveObjectArgument4ECFA0,
				object: obj,
			}
		}
	}
}
