package opennox

import "github.com/opennox/opennox/v1/server"

func (s *Server) SpellDrainManaUpdate52E210(record *server.DurSpell) int32 {
	return server.SpellDrainManaUpdate52E210(record, s.spells.duration.drainManaRuntime52E210())
}
