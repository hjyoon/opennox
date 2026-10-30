package legacy

/*
#include "monster_retreat_native_545440.h"
_Static_assert(_Generic(&nox_xxx_mobActionRetreat_545440, void (*)(nox_object_t*): 1, default: 0), "native Retreat void update ABI");
*/
import "C"
import "unsafe"

//export nox_xxx_mobActionRetreat_545440
func nox_xxx_mobActionRetreat_545440(unit *C.nox_object_t) {
	GetServer().S().MonsterActionRetreat545440(asObjectS((*nox_object_t)(unit)))
}

func monsterRetreatPublicC545440() unsafe.Pointer { return C.nox_xxx_mobActionRetreat_545440 }
