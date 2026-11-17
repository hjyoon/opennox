package server

// NetCodeCachePrependNative4ECDE0 binds an external native pointer owner to the
// existing ordered algorithm. The caller owns every load and store.
func NetCodeCachePrependNative4ECDE0[Entry comparable](entry Entry,
	storePrev func(Entry, Entry), loadFirst func() Entry,
	storeNext func(Entry, Entry), storeLast, storeFirst func(Entry),
) Entry {
	return netCodeCachePrepend4ECDE0(entry, netCodeCachePrependHooks4ECDE0[Entry]{
		storeEntryPrev: storePrev, loadFirst: loadFirst,
		storeEntryNext: storeNext, storeLast: storeLast, storeFirst: storeFirst,
	})
}

// NetCodeCacheRemoveNative4ECE10 retains independent link reloads when native
// list and entry fields overlap. It leaves allocation and lifetime to the caller.
func NetCodeCacheRemoveNative4ECE10[Entry comparable](entry Entry,
	loadNext, loadPrev func(Entry) Entry, storePrev func(Entry, Entry),
	storeLast func(Entry), storeNext func(Entry, Entry), storeFirst func(Entry),
) Entry {
	return netCodeCacheRemove4ECE10(entry, netCodeCacheRemoveHooks4ECE10[Entry]{
		loadEntryNext: loadNext, loadEntryPrev: loadPrev, storeEntryPrev: storePrev,
		storeLast: storeLast, storeEntryNext: storeNext, storeFirst: storeFirst,
	})
}
