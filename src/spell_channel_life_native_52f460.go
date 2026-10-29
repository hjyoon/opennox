package opennox

import (
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func (s *Server) SpellChannelLifeUpdate52F460(record *server.DurSpell) int32 {
	return s.S().SpellChannelLifeUpdate52F460(record, server.SpellChannelLifeRuntime52F460{
		AddMana: func(target *server.Object, amount int16) {
			legacy.Nox_xxx_playerManaAdd_4EEB80(target, int(amount))
		},
		ClearDamage: func(target *server.Object, amount int32) {
			legacy.Nox_xxx_unitDamageClear_4EE5E0(target, int(amount))
		},
		Coefficient: func(index uint32) float64 {
			return s.Balance.FloatInd("ChannelLifeCoeff", int(int32(index)))
		},
	})
}
