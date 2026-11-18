package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

var monsterInversionEntry5408D0 = func(unit *server.Object) bool {
	return GetServer().S().MonsterCastInversion5408D0(unit)
}

//export nox_xxx_mobCastInversion_5408D0
func nox_xxx_mobCastInversion_5408D0(unit *nox_object_t) C.int {
	return C.int(bool2int(monsterInversionEntry5408D0(asObjectS(unit))))
}

//export nox_xxx_unitIsMagicMissile_540B60
func nox_xxx_unitIsMagicMissile_540B60(missile, target *nox_object_t) *nox_object_t {
	return asObjectC(server.MonsterInversionMissile540B60(asObjectS(missile), asObjectS(target)))
}
