package opennox

import "github.com/opennox/opennox/v1/server"

func (s *Server) SpellTurnUndeadCreate531310(record *server.DurSpell) int32 {
	return server.SpellTurnUndeadCreate531310(record, s.spells.duration.turnUndeadRuntime531310())
}

func (s *Server) SpellTurnUndeadDestroy531420(record *server.DurSpell) {
	server.SpellTurnUndeadDestroy531420(record, s.spells.duration.turnUndeadRuntime531310())
}
