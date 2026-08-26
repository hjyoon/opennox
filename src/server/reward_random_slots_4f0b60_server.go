package server

// RewardRandomSlots4F0B60 binds GAME.EXE 004F0B60 to the server logic RNG.
// Fixed endpoint stages retain the original no-RNG short circuit.
//
//go:noinline
func (s *Server) RewardRandomSlots4F0B60(stage uint32) uint32 {
	return rewardRandomSlots4F0B60(stage, func(minimum, maximum int32) int32 {
		return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
	})
}
