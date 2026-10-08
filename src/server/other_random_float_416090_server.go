package server

// RandomFloat416090 preserves the original other stream's binary32 bounds,
// retained gameplay result and unconditional one-word advance.
func (s *Server) RandomFloat416090(min, max float32) float64 {
	return otherRandomFloat416090(s.Rand.Other, min, max)
}
