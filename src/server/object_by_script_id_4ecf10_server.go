package server

// ObjectByScriptID4ECF10 binds the portable 004ECF10 search contract to the
// native Object lists without converting object identities to legacy integers.
func (s *Server) ObjectByScriptID4ECF10(scriptID int32) *Object {
	return objectByScriptID4ECF10(objectByScriptIDHooks4ECF10[*Object]{
		firstActive: s.Objs.First,
		loadScriptIDArg: func() int32 {
			return scriptID
		},
		nextActive: func(obj *Object) *Object {
			return obj.ObjNext
		},
		firstInventory: func(obj *Object) *Object {
			return obj.InvFirstItem
		},
		nextInventory: func(obj *Object) *Object {
			return obj.InvNextItem
		},
		firstPending: func() *Object {
			return s.Objs.Pending
		},
		nextPending: func(obj *Object) *Object {
			return obj.ObjNext
		},
		firstMissile: func() *Object {
			return s.Objs.MissileList
		},
		nextMissile: func(obj *Object) *Object {
			return obj.ObjNext
		},
		loadFlagsLow: func(obj *Object) uint8 {
			return uint8(obj.ObjFlags)
		},
		loadScriptID: func(obj *Object) int32 {
			return obj.ScriptIDVal
		},
	})
}
