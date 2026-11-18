package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

var monsterHealSelection5411A0 server.MonsterHealSelection5411A0

//export nox_xxx_mobHealSomeone_5411A0
func nox_xxx_mobHealSomeone_5411A0(caster *nox_object_t) C.int {
	return C.int(GetServer().S().MonsterHealWithSelection5411A0(asObjectS(caster), &monsterHealSelection5411A0))
}

//export nox_xxx_mobMayHealThis_5412A0
func nox_xxx_mobMayHealThis_5412A0(candidate, caster *nox_object_t) {
	GetServer().S().MonsterHealCandidate5412A0(asObjectS(candidate), asObjectS(caster), &monsterHealSelection5411A0)
}
