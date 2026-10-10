package legacy

import "github.com/opennox/opennox/v1/server"

// These services retain native Object pointers across the scoring CGo entries.
// GameDataLimit preserves the original WORD return, including bit 15; the
// shared death core performs the original second read and signed comparison.
func playerDieCompetitiveRuntime54D2B0() server.PlayerDieCompetitiveRuntime54D2B0 {
	return server.PlayerDieCompetitiveRuntime54D2B0{
		Arena:       Nox_xxx_playerUpdateScoreNative54D980,
		Kotr:        Nox_xxx_playerHandleKotrDeathNative54DC40,
		Elimination: Nox_xxx_playerHandleElimDeathNative54D7A0,
		GameDataLimit: func(mode uint16) uint16 {
			return uint16(Nox_xxx_servGamedataGet_40A020(mode))
		},
		RemoveSpawned: Nox_xxx_playerRemoveSpawnedStuff_4E5AD0,
	}
}
