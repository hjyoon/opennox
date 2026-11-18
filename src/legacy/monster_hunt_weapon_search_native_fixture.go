package legacy

/*
#include "GAME4_3.h"

typedef int* (*monster_hunt_fn)(nox_object_t*);
typedef nox_object_t* (*weapon_search_fn)(nox_object_t*, float);
typedef void (*weapon_candidate_fn)(nox_object_t*, nox_object_t*);
_Static_assert(_Generic(&nox_xxx_mobActionHunt_5449D0, monster_hunt_fn:1, default:0), "hunt ABI");
_Static_assert(_Generic(&sub_544AE0, weapon_search_fn:1, default:0), "weapon search ABI");
_Static_assert(_Generic(&sub_544B20, weapon_candidate_fn:1, default:0), "weapon candidate ABI");

static monster_hunt_fn volatile hunt_indirect = nox_xxx_mobActionHunt_5449D0;
static weapon_search_fn volatile weapon_search_indirect = sub_544AE0;
static weapon_candidate_fn volatile weapon_candidate_indirect = sub_544B20;

static int* call_hunt_native(int indirect, nox_object_t* unit) {
	return indirect ? hunt_indirect(unit) : nox_xxx_mobActionHunt_5449D0(unit);
}

static nox_object_t* call_weapon_search_native(int indirect, nox_object_t* unit, float radius) {
	return indirect ? weapon_search_indirect(unit, radius) : sub_544AE0(unit, radius);
}

static void call_weapon_candidate_native(int indirect, nox_object_t* item, nox_object_t* unit) {
	if (indirect) {
		weapon_candidate_indirect(item, unit);
	} else {
		sub_544B20(item, unit);
	}
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func callMonsterHuntNative5449D0(route int, unit *server.Object) *server.AIStackItem {
	return (*server.AIStackItem)(unsafe.Pointer(C.call_hunt_native(C.int(route), asObjectC(unit))))
}

func callMonsterWeaponSearchNative544AE0(route int, unit *server.Object, radius float32) *server.Object {
	return asObjectS(C.call_weapon_search_native(C.int(route), asObjectC(unit), C.float(radius)))
}

func callMonsterWeaponCandidateNative544B20(route int, candidate, unit *server.Object) {
	C.call_weapon_candidate_native(C.int(route), asObjectC(candidate), asObjectC(unit))
}
