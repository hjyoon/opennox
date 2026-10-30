package legacy

/*
#include "monster_dying_end_native_544ca0.h"
_Static_assert(_Generic(&nox_xxx_zombieBurnDeleteCheck_544CA0, void (*)(nox_object_t*): 1, default: 0), "native dying end ABI");
*/
import "C"
import "unsafe"

//export nox_xxx_zombieBurnDeleteCheck_544CA0
func nox_xxx_zombieBurnDeleteCheck_544CA0(unit *C.nox_object_t) {
	GetServer().S().MonsterActionDyingEnd544CA0(asObjectS((*nox_object_t)(unit)), monsterActionDyingRuntime544C40())
}

func monsterDyingEndPublicC544CA0() unsafe.Pointer {
	return C.nox_xxx_zombieBurnDeleteCheck_544CA0
}
