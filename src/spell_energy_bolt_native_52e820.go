package opennox

import "github.com/opennox/opennox/v1/server"

func (s *Server) SpellEnergyBoltCreate52E820(record *server.DurSpell) int32 {
	return server.SpellEnergyBoltCreate52E820(record, s.spells.duration.energyBoltRuntime52E820())
}

func (s *Server) SpellEnergyBoltUpdate52E850(record *server.DurSpell) int32 {
	return server.SpellEnergyBoltUpdate52E850(record, s.spells.duration.energyBoltRuntime52E820())
}
