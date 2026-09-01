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

// CoopAbilityStatePending4FC680 preserves the nonzero test used by GAME.EXE
// 004FC680 before looking up a player unit.
func (s *Server) CoopAbilityStatePending4FC680() bool {
	return s.CoopAbilityState4FC670() != 0
}

// ClearCoopAbilityState4FC680 preserves the direct zero store performed by
// GAME.EXE 004FC680 after queued ability execution succeeds.
func (s *Server) ClearCoopAbilityState4FC680() {
	s.coopAbilityState4FC670 = 0
}
