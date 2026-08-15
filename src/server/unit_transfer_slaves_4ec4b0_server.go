package server

func unitTransferSlavesNative4EC4B0(source *Object, setOwner func(*Object, *Object)) {
	unitTransferSlaves4EC4B0(source, unitTransferSlavesHooks4EC4B0[*Object]{
		loadFirstOwned: func(source *Object) *Object {
			return source.Field129
		},
		loadNextOwned: func(child *Object) *Object {
			return child.Field128
		},
		loadOwner: func(source *Object) *Object {
			return source.ObjOwner
		},
		setOwner: setOwner,
	})
}

func (s *Server) unitTransferSlaves4EC4B0(source *Object) {
	unitTransferSlavesNative4EC4B0(source, s.ObjSetOwner)
}

// ObjTransferSlaves transfers source's owned-object list to source's live
// owner according to GAME.EXE 004EC4B0.
func (s *Server) ObjTransferSlaves(source *Object) {
	s.unitTransferSlaves4EC4B0(source)
}
