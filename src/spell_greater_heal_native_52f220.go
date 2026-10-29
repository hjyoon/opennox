package opennox

import (
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func (s *Server) SpellGreaterHealCreate52F220(record *server.DurSpell) int32 {
	return s.S().SpellGreaterHealCreate52F220(record, server.SpellGreaterHealRuntime52F220{
		AdjustHP: func(target *server.Object, amount int32) {
			legacy.Nox_xxx_unitAdjustHP_4EE460(target, int(amount))
		},
	})
}

func (s *Server) SpellGreaterHealUpdate52F2E0(record *server.DurSpell) int32 {
	return s.S().SpellGreaterHealUpdate52F2E0(record, server.SpellGreaterHealRuntime52F2E0{
		AdjustHP: func(target *server.Object, amount int32) {
			legacy.Nox_xxx_unitAdjustHP_4EE460(target, int(amount))
		},
		ManaSub: func(caster *server.Object, amount int32) {
			legacy.Nox_xxx_playerManaSub_4EEBF0(caster, int(amount))
		},
	})
}
