package server

// objectByExtent4ED020 binds the portable 004ED020 search contract to the
// native active-object list without converting object identities to integers.
func (s *serverObjects) objectByExtent4ED020(extent uint32) *Object {
	return objectByExtent4ED020(objectByExtentHooks4ED020[*Object]{
		first: s.First,
		loadExtentArg: func() uint32 {
			return extent
		},
		loadFlagsLow: func(obj *Object) uint8 {
			return uint8(obj.ObjFlags)
		},
		loadExtent: func(obj *Object) uint32 {
			return obj.Extent
		},
		next: func(obj *Object) *Object {
			return obj.ObjNext
		},
	})
}

// ObjectByExtent4ED020 returns the live active object whose full unsigned
// 32-bit Extent matches extent.
func (s *Server) ObjectByExtent4ED020(extent uint32) *Object {
	return s.Objs.objectByExtent4ED020(extent)
}
