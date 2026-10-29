package opennox

import "github.com/opennox/opennox/v1/server"

func (s *Server) SpellManaBombCreate530F90(record *server.DurSpell) int32 {
	return server.SpellManaBombCreate530F90(record, s.spells.duration.manaBombRuntime530F90())
}

func (s *Server) SpellManaBombUpdate5310C0(record *server.DurSpell) int32 {
	return server.SpellManaBombUpdate5310C0(record, s.spells.duration.manaBombRuntime530F90())
}

func (s *Server) SpellManaBombDestroy531290(record *server.DurSpell) {
	server.SpellManaBombDestroy531290(record, s.spells.duration.manaBombRuntime530F90())
}
