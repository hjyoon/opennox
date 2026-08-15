package server

const objectNetCodeCacheCapacity4ECCB0 = 16

// objectNetCodeCache4ECCB0 is the pointer-width-safe runtime representation
// of the original sixteen-entry object cache. Entry zero is most recently
// used; the last live entry is replaced when the cache is full.
type objectNetCodeCache4ECCB0 struct {
	objects [objectNetCodeCacheCapacity4ECCB0]*Object
	len     int
}

func (c *objectNetCodeCache4ECCB0) lookup(code uint32) *Object {
	for i := 0; i < c.len; i++ {
		obj := c.objects[i]
		if obj.NetCode != code {
			continue
		}
		if i != 0 {
			copy(c.objects[1:i+1], c.objects[:i])
			c.objects[0] = obj
		}
		return obj
	}
	return nil
}

func (c *objectNetCodeCache4ECCB0) add(obj *Object) {
	if c.len < len(c.objects) {
		copy(c.objects[1:c.len+1], c.objects[:c.len])
		c.len++
	} else {
		copy(c.objects[1:], c.objects[:len(c.objects)-1])
	}
	c.objects[0] = obj
}

func (c *objectNetCodeCache4ECCB0) remove(obj *Object) {
	for i := 0; i < c.len; i++ {
		if c.objects[i] != obj {
			continue
		}
		copy(c.objects[i:], c.objects[i+1:c.len])
		c.len--
		c.objects[c.len] = nil
		return
	}
}

func (c *objectNetCodeCache4ECCB0) clear() {
	for i := 0; i < c.len; i++ {
		c.objects[i] = nil
	}
	c.len = 0
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
