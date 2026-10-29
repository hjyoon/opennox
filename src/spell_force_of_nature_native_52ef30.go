package opennox

import "github.com/opennox/opennox/v1/server"

func (s *Server) SpellForceOfNatureCreate52EF30(record *server.DurSpell) int32 {
	return server.SpellForceOfNatureCreate52EF30(record, s.spells.duration.forceOfNatureRuntime52EF30())
}

func (s *Server) SpellForceOfNatureUpdate52EFD0(record *server.DurSpell) int32 {
	return server.SpellForceOfNatureUpdate52EFD0(record, s.spells.duration.forceOfNatureRuntime52EF30())
}

func (s *Server) SpellForceOfNatureDestroy52F1D0(record *server.DurSpell) {
	server.SpellForceOfNatureDestroy52F1D0(record, s.spells.duration.forceOfNatureRuntime52EF30())
}
