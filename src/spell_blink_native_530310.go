package opennox

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func (s *Server) SpellBlinkCreate530310(record *server.DurSpell) int32 {
	return server.SpellBlinkCreate530310(record, s.spells.duration.blinkRuntime530310())
}

func (s *Server) SpellBlinkUpdate530380(record *server.DurSpell) int32 {
	return server.SpellBlinkUpdate530380(record, s.spells.duration.blinkRuntime530310())
}

func (s *Server) SpellTeleportCreateWake530560(target *server.Object, origin, destination *types.Pointf) {
	server.SpellTeleportCreateWake530560(target, origin, destination, s.spells.duration.blinkRuntime530310())
}
