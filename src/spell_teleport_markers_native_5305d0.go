package opennox

import "github.com/opennox/opennox/v1/server"

func (s *Server) SpellTeleportPopCreate530820(record *server.DurSpell) int32 {
	return server.SpellTeleportPopCreate530820(record, s.spells.duration.teleportPopRuntime530820())
}

func (s *Server) SpellTeleportPopUpdate530880(record *server.DurSpell) int32 {
	return server.SpellTeleportPopUpdate530880(record, s.spells.duration.teleportPopRuntime530820())
}

func (s *Server) SpellTeleportToMarkCreate5305D0(record *server.DurSpell) int32 {
	return server.SpellTeleportToMarkCreate5305D0(record, s.spells.duration.teleportToMarkRuntime5305D0())
}

func (s *Server) SpellTeleportToMarkUpdate530650(record *server.DurSpell) int32 {
	return server.SpellTeleportToMarkUpdate530650(record, s.spells.duration.teleportToMarkRuntime5305D0())
}
