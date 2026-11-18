package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/opennox/v1/server"
)

//export sub_540CE0
func sub_540CE0(unit *nox_object_t, id C.int) C.int {
	return C.int(bool2int(server.MonsterHasSpellEnchant540CE0(asObjectS(unit), spell.ID(id))))
}

//export sub_540D20
func sub_540D20(id C.int) C.int {
	return C.int(bool2int(server.MonsterSpellIsSummon540D20(spell.ID(id))))
}

//export sub_540D40
func sub_540D40(unit *nox_object_t) C.int {
	return C.int(bool2int(GetServer().S().Spells.Dur.HasCasterSummon540D40(asObjectS(unit))))
}
