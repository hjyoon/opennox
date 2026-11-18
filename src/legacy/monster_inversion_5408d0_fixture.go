package legacy

/*
#include "GAME4_3.h"

typedef int (*monster_inversion_entry)(nox_object_t*);
typedef nox_object_t* (*monster_inversion_missile)(nox_object_t*, nox_object_t*);

_Static_assert(_Generic(&nox_xxx_mobCastInversion_5408D0, monster_inversion_entry: 1, default: 0), "monster inversion ABI");
_Static_assert(_Generic(&nox_xxx_unitIsMagicMissile_540B60, monster_inversion_missile: 1, default: 0), "monster missile filter ABI");

static int monster_inversion_fixture(nox_object_t* unit) {
	monster_inversion_entry volatile cast = nox_xxx_mobCastInversion_5408D0;
	return cast(unit);
}

static nox_object_t* monster_inversion_missile_fixture(nox_object_t* missile, nox_object_t* target) {
	monster_inversion_missile volatile filter = nox_xxx_unitIsMagicMissile_540B60;
	return filter(missile, target);
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func monsterInversionFixture5408D0(unit *server.Object) int {
	return int(C.monster_inversion_fixture(asObjectC(unit)))
}

func monsterInversionMissileFixture540B60(missile, target *server.Object) *server.Object {
	return asObjectS(C.monster_inversion_missile_fixture(asObjectC(missile), asObjectC(target)))
}
