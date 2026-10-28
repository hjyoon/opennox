package opennox

import "github.com/opennox/opennox/v1/server"

func (s *Server) SpellSwapCreate530CA0(record *server.DurSpell) int32 {
	return server.SpellSwapCreate530CA0(record, s.spells.duration.swapRuntime530CA0())
}

func (s *Server) SpellSwapUpdate530D30(record *server.DurSpell) int32 {
	return server.SpellSwapUpdate530D30(record, s.spells.duration.swapRuntime530CA0())
}
