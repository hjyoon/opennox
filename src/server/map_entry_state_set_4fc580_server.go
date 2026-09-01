package server

// MapEntryState4FC580 returns the complete native map-entry state dword.
func (s *Server) MapEntryState4FC580() int32 {
	return s.mapEntryState4FC580
}

// SetMapEntryState4FC580 binds GAME.EXE 004FC580 to the native server-owned
// state while preserving the original full-width store and return value.
func (s *Server) SetMapEntryState4FC580(value int32) int32 {
	return mapEntryStateSet4FC580(mapEntryStateSetHooks4FC580{
		loadValueArg: func() int32 {
			return value
		},
		storeValue: func(value int32) {
			s.mapEntryState4FC580 = value
		},
	})
}

// MapEntryStatePending4FC600 preserves the nonzero test used by GAME.EXE
// 004FC600 before dispatching the MapEntry event.
func (s *Server) MapEntryStatePending4FC600() bool {
	return s.MapEntryState4FC580() != 0
}

// MapEntryStateRequestsPlayerInit4FC6D0 preserves the exact comparison with
// one used by GAME.EXE 004FC6D0 before initializing player units.
func (s *Server) MapEntryStateRequestsPlayerInit4FC6D0() bool {
	return s.MapEntryState4FC580() == 1
}
