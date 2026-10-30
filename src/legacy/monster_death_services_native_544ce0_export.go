package legacy

/*
#include "monster_death_services_native_544ce0.h"
_Static_assert(_Generic(&nox_xxx_zombieBurnDelete_544CE0, void (*)(nox_object_t*): 1, default: 0), "native burn delete void ABI");
_Static_assert(_Generic(&nox_xxx_createReleasedSoul_544E60, void (*)(nox_object_t*): 1, default: 0), "native released soul void ABI");
*/
import "C"
import "unsafe"

import "github.com/opennox/opennox/v1/server"

//export nox_xxx_zombieBurnDelete_544CE0
func nox_xxx_zombieBurnDelete_544CE0(unit *C.nox_object_t) {
	zombieBurnDeleteNative544CE0(asObjectS((*nox_object_t)(unit)))
}

//export nox_xxx_createReleasedSoul_544E60
func nox_xxx_createReleasedSoul_544E60(unit *C.nox_object_t) {
	server.MonsterCreateReleasedSoul544E60(asObjectS((*nox_object_t)(unit)), monsterReleasedSoulRuntime544E60())
}

func monsterBurnDeletePublicC544CE0() unsafe.Pointer   { return C.nox_xxx_zombieBurnDelete_544CE0 }
func monsterReleasedSoulPublicC544E60() unsafe.Pointer { return C.nox_xxx_createReleasedSoul_544E60 }
