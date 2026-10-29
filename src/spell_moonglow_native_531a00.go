package opennox

import "github.com/opennox/opennox/v1/server"

func (s *Server) SpellMoonglowCreate531A00(record *server.DurSpell) int32 {
	return server.SpellMoonglowCreate531A00(record, s.spells.duration.moonglowRuntime531A00())
}

func (s *Server) SpellMoonglowDestroy531AF0(record *server.DurSpell) {
	server.SpellMoonglowDestroy531AF0(record, s.spells.duration.moonglowRuntime531A00())
	// Match ordinary duration dispatch's lifetime cleanup even when the
	// target vanished before destruction and the original root did no work.
	delete(s.spells.duration.moonglowVisuals, record)
}
