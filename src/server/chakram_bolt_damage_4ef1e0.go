package server

import noxflags "github.com/opennox/opennox/v1/common/flags"

const chakramArcherBoltTypeName4EF1E0 = "ArcherBolt"

// chakramCalcBoltDamage4EF1E0 preserves GAME.EXE 004EF1E0. The modifier's
// binary32 coefficient is widened before multiplication. Cooperative
// ArcherBolt damage replaces the modifier minimum with BoltSoloDamageMin;
// every other case uses the modifier's fixed minimum.
func chakramCalcBoltDamage4EF1E0(
	strength int32,
	modifier *Modifier,
	cooperative bool,
	archerBoltType uint32,
	boltSoloDamageMin float64,
) float64 {
	delta := strength - int32(modifier.ReqStrength60)
	scaled := float64(delta) * float64(modifier.DamageCoeffOrArmor64)
	if cooperative && modifier.TypeInd == archerBoltType {
		return boltSoloDamageMin + scaled
	}
	return scaled + float64(modifier.DamageMin72)
}

func (s *Server) chakramCalcBoltDamageNative4EF1E0(strength int32, modifier *Modifier) float32 {
	return float32(chakramCalcBoltDamage4EF1E0(
		strength,
		modifier,
		noxflags.HasGame(noxflags.GameModeCoop),
		uint32(s.Types.IndByID(chakramArcherBoltTypeName4EF1E0)),
		s.Balance.Float("BoltSoloDamageMin"),
	))
}
