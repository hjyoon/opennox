package opennox

import "github.com/opennox/opennox/v1/server"

func (s *Server) SpellPlasmaCreate531580(record *server.DurSpell) int32 {
	return server.SpellPlasmaCreate531580(record, s.spells.duration.plasmaRuntime531580())
}

func (s *Server) SpellPlasmaUpdate531600(record *server.DurSpell) int32 {
	return server.SpellPlasmaUpdate531600(record, s.spells.duration.plasmaRuntime531580())
}

func (s *Server) SpellPlasmaDestroy5319E0(record *server.DurSpell) {
	server.SpellPlasmaDestroy5319E0(record, s.spells.duration.plasmaRuntime531580())
}
