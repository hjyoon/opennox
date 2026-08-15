package server

const objectNetCodeCacheCapacity4ECCB0 = 16

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

func (list *objectNetCodeCacheList4ECD90) remove(entry *objectNetCodeCacheEntry4ECD90) {
	if entry.next != nil {
		entry.next.prev = entry.prev
	} else {
		list.last = entry.prev
	}
	if entry.prev != nil {
		entry.prev.next = entry.next
	} else {
		list.first = entry.next
	}
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

func (c *objectNetCodeCache4ECCB0) init() {
	c.free = objectNetCodeCacheList4ECD90{}
	c.used = objectNetCodeCacheList4ECD90{}
	for i := range c.entries {
		c.free.prepend(&c.entries[i])
	}
	c.initialized = true
}

func (c *objectNetCodeCache4ECCB0) lookup(code uint32) *Object {
	return netCodeCacheLookupObject4ECD90(code, netCodeCacheLookupHooks4ECD90[
		*objectNetCodeCacheEntry4ECD90,
		*Object,
	]{
		loadNeedsInit: func() bool {
			return !c.initialized
		},
		initCache: c.init,
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
		removeUsed: c.used.remove,
		prependUsed: func(entry *objectNetCodeCacheEntry4ECD90) {
			c.used.prepend(entry)
		},
	})
}

func (c *objectNetCodeCache4ECCB0) nextUnused() *objectNetCodeCacheEntry4ECD90 {
	entry := c.free.first
	if entry != nil {
		c.free.first = entry.next
	}
	return entry
}

func (c *objectNetCodeCache4ECCB0) add(obj *Object) {
	entry := c.nextUnused()
	if entry == nil {
		entry = c.used.last
		entry.object = obj
		c.used.remove(entry)
		c.used.prepend(entry)
		return
	}
	entry.object = obj
	c.used.prepend(entry)
}

func (c *objectNetCodeCache4ECCB0) remove(obj *Object) {
	if !c.initialized {
		return
	}
	for entry := c.used.first; entry != nil; entry = entry.next {
		if entry.object != obj {
			continue
		}
		c.used.remove(entry)
		c.free.prepend(entry)
		entry.object = nil
		return
	}
}

func (c *objectNetCodeCache4ECCB0) clear() {
	if !c.initialized {
		return
	}
	for entry := c.used.first; entry != nil; {
		next := entry.next
		c.used.remove(entry)
		c.free.prepend(entry)
		entry.object = nil
		entry = next
	}
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
		cacheAdd:    s.Objs.netCodeCache.add,
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

// ObjectNetCodeCacheRemove4ECFA0 keeps deletion from leaving a stale pointer
// in the pointer-width-safe cache. The exact standalone 004ECFA0 routine is a
// later sequential oracle range; this method preserves its runtime consumer.
func (s *Server) ObjectNetCodeCacheRemove4ECFA0(obj *Object) {
	s.Objs.netCodeCache.remove(obj)
}

// ObjectNetCodeCacheClear4ECFE0 returns all runtime cache entries to the empty
// state without retaining native object pointers.
func (s *Server) ObjectNetCodeCacheClear4ECFE0() {
	s.Objs.netCodeCache.clear()
}
