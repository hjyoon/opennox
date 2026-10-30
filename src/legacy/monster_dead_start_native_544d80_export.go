package legacy

/*
#include "monster_dead_start_native_544d80.h"
_Static_assert(_Generic(&nox_xxx_mobActionDead1_544D80, char (*)(nox_object_t*): 1, default: 0), "native dead start ABI");
*/
import "C"
import "unsafe"

//export nox_xxx_mobActionDead1_544D80
func nox_xxx_mobActionDead1_544D80(unit *C.nox_object_t) C.char {
	u := asObjectS((*nox_object_t)(unit))
	GetServer().S().MonsterActionDeadStart544D80(u, monsterActionDeadRuntime544D80())
	return C.char(uint8(u.ObjFlags))
}

func monsterDeadStartPublicC544D80() unsafe.Pointer {
	return C.nox_xxx_mobActionDead1_544D80
}
