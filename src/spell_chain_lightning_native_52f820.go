package opennox

import "github.com/opennox/opennox/v1/server"

// Both native C callbacks and the duration dispatcher use the same runtime,
// including the native wand sidecar and real ray allocation/network services.
func (s *Server) SpellChainLightningCreate52F820(record *server.DurSpell) int32 {
	return server.SpellChainLightningCreate52F820(record, s.spells.duration.chainLightningRuntime52F820())
}

func (s *Server) SpellChainLightningUpdate52F8A0(record *server.DurSpell) int32 {
	return server.SpellChainLightningUpdate52F8A0(record, s.spells.duration.chainLightningRuntime52F820())
}

func (s *Server) SpellChainLightningDestroy530100(record *server.DurSpell) uint8 {
	return server.SpellChainLightningDestroy530100(record, s.spells.duration.chainLightningRuntime52F820())
}
