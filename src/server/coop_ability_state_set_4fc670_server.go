package server

// CoopAbilityState4FC670 returns the complete native queued
// cooperative-ability state dword.
func (s *Server) CoopAbilityState4FC670() int32 {
	return s.coopAbilityState4FC670
}

// SetCoopAbilityState4FC670 binds GAME.EXE 004FC670 to the native
// server-owned state while preserving the original full-width store and
// return value.
func (s *Server) SetCoopAbilityState4FC670(value int32) int32 {
	return coopAbilityStateSet4FC670(coopAbilityStateSetHooks4FC670{
		loadValueArg: func() int32 {
			return value
		},
		storeValue: func(value int32) {
			s.coopAbilityState4FC670 = value
		},
	})
}
