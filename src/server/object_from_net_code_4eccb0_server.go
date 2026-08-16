package server

const objectNetCodeCacheCapacity4ECCB0 = netCodeCacheInitArrayCapacity4ECE50

type objectNetCodeCacheEntry4ECD90 struct {
	object *Object
	prev   *objectNetCodeCacheEntry4ECD90
	next   *objectNetCodeCacheEntry4ECD90
}

type objectNetCodeCacheList4ECD90 struct {
	first *objectNetCodeCacheEntry4ECD90
	last  *objectNetCodeCacheEntry4ECD90
}

func (list *objectNetCodeCacheList4ECD90) prepend(entry *objectNetCodeCacheEntry4ECD90) *objectNetCodeCacheEntry4ECD90 {
	return netCodeCachePrepend4ECDE0(entry, netCodeCachePrependHooks4ECDE0[*objectNetCodeCacheEntry4ECD90]{
		storeEntryPrev: func(entry, prev *objectNetCodeCacheEntry4ECD90) {
			entry.prev = prev
		},
		loadFirst: func() *objectNetCodeCacheEntry4ECD90 {
			return list.first
		},
		storeEntryNext: func(entry, next *objectNetCodeCacheEntry4ECD90) {
			entry.next = next
		},
		storeLast: func(entry *objectNetCodeCacheEntry4ECD90) {
			list.last = entry
		},
		storeFirst: func(entry *objectNetCodeCacheEntry4ECD90) {
			list.first = entry
		},
	})
}

func (list *objectNetCodeCacheList4ECD90) remove(entry *objectNetCodeCacheEntry4ECD90) *objectNetCodeCacheEntry4ECD90 {
	return netCodeCacheRemove4ECE10(entry, netCodeCacheRemoveHooks4ECE10[*objectNetCodeCacheEntry4ECD90]{
		loadEntryNext: func(entry *objectNetCodeCacheEntry4ECD90) *objectNetCodeCacheEntry4ECD90 {
			return entry.next
		},
		loadEntryPrev: func(entry *objectNetCodeCacheEntry4ECD90) *objectNetCodeCacheEntry4ECD90 {
			return entry.prev
		},
		storeEntryPrev: func(entry, prev *objectNetCodeCacheEntry4ECD90) {
			entry.prev = prev
		},
		storeLast: func(entry *objectNetCodeCacheEntry4ECD90) {
			list.last = entry
		},
		storeEntryNext: func(entry, next *objectNetCodeCacheEntry4ECD90) {
			entry.next = next
		},
		storeFirst: func(entry *objectNetCodeCacheEntry4ECD90) {
			list.first = entry
		},
	})
}

// objectNetCodeCache4ECCB0 is the pointer-width-safe runtime representation
// of the original sixteen-entry free and used lists. The standalone list
// helper ranges following 004ECD90 remain separate sequential audit units.
type objectNetCodeCache4ECCB0 struct {
	free        objectNetCodeCacheList4ECD90
	entries     [objectNetCodeCacheCapacity4ECCB0]objectNetCodeCacheEntry4ECD90
	used        objectNetCodeCacheList4ECD90
	initialized bool
}

func (c *objectNetCodeCache4ECCB0) init() *objectNetCodeCacheEntry4ECD90 {
	var entries [netCodeCacheInitArrayCapacity4ECE50]*objectNetCodeCacheEntry4ECD90
	for i := range entries {
		entries[i] = &c.entries[i]
	}
	return netCodeCacheInitArray4ECE50(entries, netCodeCacheInitArrayHooks4ECE50[
		*objectNetCodeCacheEntry4ECD90,
		*objectNetCodeCacheEntry4ECD90,
	]{
		storeUsedFirst: func(entry *objectNetCodeCacheEntry4ECD90) {
			c.used.first = entry
		},
		storeUsedLast: func(entry *objectNetCodeCacheEntry4ECD90) {
			c.used.last = entry
		},
		storeFreeFirst: func(entry *objectNetCodeCacheEntry4ECD90) {
			c.free.first = entry
		},
		storeFreeLast: func(entry *objectNetCodeCacheEntry4ECD90) {
			c.free.last = entry
		},
		prependFree: c.free.prepend,
		clearNeedsInit: func() {
			c.initialized = true
		},
	})
}

func (c *objectNetCodeCache4ECCB0) lookup(code uint32) *Object {
	return netCodeCacheLookupObject4ECD90(code, netCodeCacheLookupHooks4ECD90[
		*objectNetCodeCacheEntry4ECD90,
		*Object,
	]{
		loadNeedsInit: func() bool {
			return !c.initialized
		},
		initCache: func() {
			c.init()
		},
		loadFirstUsed: func() *objectNetCodeCacheEntry4ECD90 {
			return c.used.first
		},
		loadObject: func(entry *objectNetCodeCacheEntry4ECD90) *Object {
			return entry.object
		},
		loadNetCode: func(obj *Object) uint32 {
			return obj.NetCode
		},
		loadNext: func(entry *objectNetCodeCacheEntry4ECD90) *objectNetCodeCacheEntry4ECD90 {
			return entry.next
		},
		removeUsed: func(entry *objectNetCodeCacheEntry4ECD90) {
			c.used.remove(entry)
		},
		prependUsed: func(entry *objectNetCodeCacheEntry4ECD90) {
			c.used.prepend(entry)
		},
	})
}

func (c *objectNetCodeCache4ECCB0) nextUnused() *objectNetCodeCacheEntry4ECD90 {
	return netCodeCacheNextUnused4ECEF0(netCodeCacheNextUnusedHooks4ECEF0[*objectNetCodeCacheEntry4ECD90]{
		loadFirstFree: func() *objectNetCodeCacheEntry4ECD90 {
			return c.free.first
		},
		loadEntryNext: func(entry *objectNetCodeCacheEntry4ECD90) *objectNetCodeCacheEntry4ECD90 {
			return entry.next
		},
		storeFirstFree: func(entry *objectNetCodeCacheEntry4ECD90) {
			c.free.first = entry
		},
	})
}

func (c *objectNetCodeCache4ECCB0) add(obj *Object) *objectNetCodeCacheEntry4ECD90 {
	return netCodeCacheAddObject4ECEA0(netCodeCacheAddHooks4ECEA0[
		*objectNetCodeCacheEntry4ECD90,
		*Object,
		*objectNetCodeCacheEntry4ECD90,
	]{
		nextUnused: c.nextUnused,
		loadObject: func() *Object {
			return obj
		},
		loadLastUsed: func() *objectNetCodeCacheEntry4ECD90 {
			return c.used.last
		},
		storeObject: func(entry *objectNetCodeCacheEntry4ECD90, obj *Object) {
			entry.object = obj
		},
		removeUsed: func(entry *objectNetCodeCacheEntry4ECD90) {
			c.used.remove(entry)
		},
		prependUsed: c.used.prepend,
	})
}

func (c *objectNetCodeCache4ECCB0) remove(obj *Object) netCodeCacheRemoveObjectResult4ECFA0[*objectNetCodeCacheEntry4ECD90, *Object] {
	return netCodeCacheRemoveObject4ECFA0(netCodeCacheRemoveObjectHooks4ECFA0[
		*objectNetCodeCacheEntry4ECD90,
		*Object,
	]{
		loadNeedsInit: func() uint32 {
			if c.initialized {
				return 0
			}
			return 1
		},
		loadFirstUsed: func() *objectNetCodeCacheEntry4ECD90 {
			return c.used.first
		},
		loadObjectArg: func() *Object {
			return obj
		},
		loadEntryObject: func(entry *objectNetCodeCacheEntry4ECD90) *Object {
			return entry.object
		},
		loadEntryNext: func(entry *objectNetCodeCacheEntry4ECD90) *objectNetCodeCacheEntry4ECD90 {
			return entry.next
		},
		removeUsed: func(entry *objectNetCodeCacheEntry4ECD90) {
			c.used.remove(entry)
		},
		prependFree: c.free.prepend,
	})
}

func (c *objectNetCodeCache4ECCB0) clear() netCodeCacheClearResult4ECFE0[*objectNetCodeCacheEntry4ECD90] {
	return netCodeCacheClear4ECFE0(netCodeCacheClearHooks4ECFE0[*objectNetCodeCacheEntry4ECD90]{
		loadNeedsInit: func() uint32 {
			if c.initialized {
				return 0
			}
			return 1
		},
		loadFirstUsed: func() *objectNetCodeCacheEntry4ECD90 {
			return c.used.first
		},
		loadEntryNext: func(entry *objectNetCodeCacheEntry4ECD90) *objectNetCodeCacheEntry4ECD90 {
			return entry.next
		},
		removeUsed: func(entry *objectNetCodeCacheEntry4ECD90) {
			c.used.remove(entry)
		},
		prependFree: c.free.prepend,
	})
}

func (c *objectNetCodeCache4ECCB0) usedLen() int {
	n := 0
	for entry := c.used.first; entry != nil; entry = entry.next {
		n++
	}
	return n
}

// ObjectFromNetCode4ECCB0 binds the portable 004ECCB0 search contract to
// native Object and Player pointers without converting either to an integer.
func (s *Server) ObjectFromNetCode4ECCB0(code uint32) *Object {
	return objectFromNetCode4ECCB0(code, objectFromNetCodeHooks4ECCB0[*Object, *Player]{
		cacheLookup: s.Objs.netCodeCache.lookup,
		cacheAdd: func(obj *Object) {
			s.Objs.netCodeCache.add(obj)
		},
		firstObject: s.Objs.First,
		nextObject: func(obj *Object) *Object {
			return obj.ObjNext
		},
		firstItem: func(obj *Object) *Object {
			return obj.InvFirstItem
		},
		nextItem: func(obj *Object) *Object {
			return obj.InvNextItem
		},
		firstPending: func() *Object {
			return s.Objs.Pending
		},
		nextPending: func(obj *Object) *Object {
			return obj.ObjNext
		},
		firstPlayer: s.Players.First,
		nextPlayer:  s.Players.Next,
		loadPlayerUnit: func(player *Player) *Object {
			return player.PlayerUnit
		},
		loadFlagsLow: func(obj *Object) uint8 {
			return uint8(obj.ObjFlags)
		},
		loadNetCode: func(obj *Object) uint32 {
			return obj.NetCode
		},
	})
}

// ObjectNetCodeCacheRemove4ECFA0 binds the exact standalone 004ECFA0 contract
// to the pointer-width-safe cache. The sole original caller ignores EAX, so
// this runtime boundary deliberately discards the contract's mixed-domain
// result after preserving all observable list and entry operations.
func (s *Server) ObjectNetCodeCacheRemove4ECFA0(obj *Object) {
	s.Objs.netCodeCache.remove(obj)
}

// ObjectNetCodeCacheClear4ECFE0 returns every used runtime cache entry to the
// free list. The sole original caller ignores EAX, so this boundary discards
// the mixed-domain result after preserving the original entry operations.
func (s *Server) ObjectNetCodeCacheClear4ECFE0() {
	s.Objs.netCodeCache.clear()
}
