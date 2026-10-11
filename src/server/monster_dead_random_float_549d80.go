package server

// MonsterDeadRandomFloat549D80 preserves the retained 00416030 result and
// the death callers' binary32 FSTPS boundary (for example 00549F20/00549F4B).
// The shared narrowing helper implements the gameplay x87 round-toward-zero
// store locally; neither the thread's FPU mode nor RNG consumption changes.
func (s *Server) MonsterDeadRandomFloat549D80(min, max float32) float32 {
	return float32(monsterMoveToRunSpill544440(s.RandomFloat416030(min, max)))
}
