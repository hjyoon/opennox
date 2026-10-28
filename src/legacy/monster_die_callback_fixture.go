package legacy

/*
#include "defs.h"
#include "GAME5.h"
_Static_assert(_Generic(&sub_54A7D0, int (*)(nox_object_t*): 1, default: 0), "Swordsman death native signature");
_Static_assert(_Generic(&sub_54A850, int (*)(nox_object_t*): 1, default: 0), "Shaman death native signature");
_Static_assert(_Generic(&sub_54A890, int (*)(nox_object_t*): 1, default: 0), "Archer death native signature");
_Static_assert(_Generic(&sub_54A900, int (*)(nox_object_t*): 1, default: 0), "Ogre death native signature");
_Static_assert(_Generic(&sub_54A950, int (*)(nox_object_t*): 1, default: 0), "Warlord death native signature");
static int nox_monster_die_probe(int kind, nox_object_t* unit) {
 int (*const entries[])(nox_object_t*) = {sub_54A7D0, sub_54A850, sub_54A890, sub_54A900, sub_54A950};
 return entries[kind-1](unit);
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func monsterDieCallbackCProbe54A7D0(kind server.MonsterDieCallbackKind54A7D0, unit *server.Object) int32 {
	return int32(C.nox_monster_die_probe(C.int(kind), asObjectC(unit)))
}
