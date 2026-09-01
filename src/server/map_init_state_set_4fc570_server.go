package server

// MapInitState4FC570 returns the complete native map-initialize state dword.
func (s *Server) MapInitState4FC570() int32 {
	return s.mapInitState4FC570
}

// SetMapInitState4FC570 binds GAME.EXE 004FC570 to the native server-owned
// state while preserving the original full-width store and return value.
func (s *Server) SetMapInitState4FC570(value int32) int32 {
	return mapInitStateSet4FC570(mapInitStateSetHooks4FC570{
		loadValueArg: func() int32 {
			return value
		},
		storeValue: func(value int32) {
			s.mapInitState4FC570 = value
		},
	})
}
