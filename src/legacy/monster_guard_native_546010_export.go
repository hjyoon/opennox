package legacy

/*
#include "monster_guard_native_546010.h"
_Static_assert(_Generic(&nox_xxx_mobActionGuard_546010, void (*)(nox_object_t*): 1, default: 0), "native Guard void update ABI");
*/
import "C"
import "unsafe"

//export nox_xxx_mobActionGuard_546010
func nox_xxx_mobActionGuard_546010(unit *C.nox_object_t) {
	GetServer().S().MonsterActionGuard546010(asObjectS((*nox_object_t)(unit)), Sub_5466F0)
}

func monsterGuardPublicC546010() unsafe.Pointer { return C.nox_xxx_mobActionGuard_546010 }
