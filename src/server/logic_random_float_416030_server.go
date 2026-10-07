package server

// RandomFloat416030 preserves the original logic stream's binary32 bounds,
// retained 53-bit gameplay result, and zero/unordered no-step branch.
func (s *Server) RandomFloat416030(min, max float32) float64 {
	return logicRandomFloat416030(s.Rand.Logic, min, max)
}
