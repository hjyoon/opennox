package legacy

/*
#include "GAME1_1.h"

typedef void (*monster_polygon_entry)(nox_object_t*);
_Static_assert(_Generic(&nox_xxx_monsterPolygonEnter_421FF0, monster_polygon_entry: 1, default: 0), "monster polygon ABI");

static void monster_polygon_fixture(nox_object_t* unit) {
	monster_polygon_entry volatile enter = nox_xxx_monsterPolygonEnter_421FF0;
	enter(unit);
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func monsterPolygonFixture421FF0(unit *server.Object) {
	C.monster_polygon_fixture(asObjectC(unit))
}
