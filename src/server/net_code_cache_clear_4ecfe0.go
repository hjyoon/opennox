package server

type netCodeCacheClearResultKind4ECFE0 uint8

const (
	netCodeCacheClearInitial4ECFE0 netCodeCacheClearResultKind4ECFE0 = iota
	netCodeCacheClearEntry4ECFE0
)

// netCodeCacheClearResult4ECFE0 keeps the original mixed EAX return domains
// distinct. GAME.EXE returns the raw initialization flag before the cache can
// be cleared, zero for an initialized empty list, and the final free-list
// prepend result after clearing a nonempty list.
type netCodeCacheClearResult4ECFE0[Entry any] struct {
	kind        netCodeCacheClearResultKind4ECFE0
	initialFlag uint32
	entry       Entry
}

// netCodeCacheClearHooks4ECFE0 separates entry identities from the original
// ABI32 list storage so that native pointers remain intact on every target.
type netCodeCacheClearHooks4ECFE0[Entry comparable] struct {
	loadNeedsInit func() uint32
	loadFirstUsed func() Entry
	loadEntryNext func(Entry) Entry
	removeUsed    func(Entry)
	prependFree   func(Entry) Entry
}

// netCodeCacheClear4ECFE0 preserves GAME.EXE 004ECFE0. Each successor is
// loaded before the current entry is removed, then that exact cached successor
// drives the next iteration. Entry object fields are neither read nor cleared.
func netCodeCacheClear4ECFE0[Entry comparable](hooks netCodeCacheClearHooks4ECFE0[Entry]) netCodeCacheClearResult4ECFE0[Entry] {
	needsInit := hooks.loadNeedsInit()
	if needsInit != 0 {
		return netCodeCacheClearResult4ECFE0[Entry]{
			kind:        netCodeCacheClearInitial4ECFE0,
			initialFlag: needsInit,
		}
	}

	entry := hooks.loadFirstUsed()
	var nilEntry Entry
	if entry == nilEntry {
		return netCodeCacheClearResult4ECFE0[Entry]{
			kind:        netCodeCacheClearInitial4ECFE0,
			initialFlag: needsInit,
		}
	}

	for {
		next := hooks.loadEntryNext(entry)
		hooks.removeUsed(entry)
		result := hooks.prependFree(entry)
		if next == nilEntry {
			return netCodeCacheClearResult4ECFE0[Entry]{
				kind:  netCodeCacheClearEntry4ECFE0,
				entry: result,
			}
		}
		entry = next
	}
}
