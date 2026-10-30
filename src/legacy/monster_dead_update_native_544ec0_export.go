package legacy

/*
#include "monster_dead_update_native_544ec0.h"
_Static_assert(_Generic(&nox_xxx_mobActionDead2_544EC0, void (*)(nox_object_t*): 1, default: 0), "native dead update ABI");
_Static_assert(_Generic(&sub_544F70, void* (*)(nox_object_t*): 1, default: 0), "native dead cleanup pointer ABI");
*/
import "C"
import "unsafe"

import "github.com/opennox/opennox/v1/server"

//export nox_xxx_mobActionDead2_544EC0
func nox_xxx_mobActionDead2_544EC0(unit *C.nox_object_t) {
	GetServer().S().MonsterActionDeadUpdate544EC0(asObjectS((*nox_object_t)(unit)), monsterActionDeadRuntime544D80())
}

//export sub_544F70
func sub_544F70(unit *C.nox_object_t) unsafe.Pointer {
	return server.MonsterDeadCleanupNative544F70(asObjectS((*nox_object_t)(unit)), monsterActionDeadRuntime544D80())
}

func monsterDeadUpdatePublicC544EC0() unsafe.Pointer  { return C.nox_xxx_mobActionDead2_544EC0 }
func monsterDeadCleanupPublicC544F70() unsafe.Pointer { return C.sub_544F70 }
