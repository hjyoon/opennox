package legacy

/*
#include "player_cancel_spells_4feae0.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func playerCancelSpellsLegacy4FEAE0(caster *server.Object) int32 {
	return GetServer().S().Spells.Dur.PlayerCancelSpells4FEAE0(caster)
}

func playerCancelSpellsExportCall4FEAE0(caster *server.Object) int32 {
	return int32(C.nox_xxx_playerCancelSpells_4FEAE0(asObjectC(caster)))
}

//export nox_xxx_playerCancelSpells_4FEAE0
func nox_xxx_playerCancelSpells_4FEAE0(caster *nox_object_t) C.int32_t {
	return C.int32_t(playerCancelSpellsLegacy4FEAE0(asObjectS(caster)))
}
