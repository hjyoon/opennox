package opennox

import "github.com/opennox/opennox/v1/server"

func (s *Server) SpellTeleportToTargetCreate530A30(record *server.DurSpell) int32 {
	return server.SpellTeleportToTargetCreate530A30(record, s.spells.duration.teleportToTargetRuntime530A30())
}

func (s *Server) SpellTeleportToTargetUpdate530B70(record *server.DurSpell) int32 {
	return server.SpellTeleportToTargetUpdate530B70(record, s.spells.duration.teleportToTargetRuntime530A30())
}
