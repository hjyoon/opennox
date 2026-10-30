package legacy

/*
#include "monster_dying_start_native_544c40.h"
#include "GAME1_1.h"
_Static_assert(_Generic(&nox_xxx_mobGenericDeath_544C40, int (*)(nox_object_t*): 1, default: 0), "native dying start int ABI");
_Static_assert(_Generic(&nox_xxx_monsterGetSoundSet_424300, void* (*)(nox_object_t*): 1, default: 0), "native monster sound pointer ABI");
*/
import "C"
import "unsafe"

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_mobGenericDeath_544C40
func nox_xxx_mobGenericDeath_544C40(unit *C.nox_object_t) C.int {
	runtime := monsterActionDyingRuntime544C40()
	die := runtime.DieFunc
	var result C.int
	runtime.DieFunc = func(fn unsafe.Pointer, unit *server.Object) {
		die(fn, unit)
		// All five admitted stock DIE_FUNCTION callers return one, even when
		// no item drops. A missing callback leaves the original zero result.
		result = 1
	}
	GetServer().S().MonsterActionDyingStart544C40(asObjectS((*nox_object_t)(unit)), runtime)
	return result
}

//export nox_xxx_monsterGetSoundSet_424300
func nox_xxx_monsterGetSoundSet_424300(unit *C.nox_object_t) unsafe.Pointer {
	u := asObjectS((*nox_object_t)(unit))
	if u == nil || uint8(u.ObjClass)&uint8(object.ClassMonster) == 0 {
		return nil
	}
	return (*server.MonsterUpdateData)(u.UpdateData).SoundSet122
}

func monsterDyingStartPublicC544C40() unsafe.Pointer { return C.nox_xxx_mobGenericDeath_544C40 }
func monsterSoundSetPublicC424300() unsafe.Pointer   { return C.nox_xxx_monsterGetSoundSet_424300 }
