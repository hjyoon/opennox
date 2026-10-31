package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

// The factory only creates lazy callbacks. A cached Mimic idle path must
// return without resolving the server or either remaining type predicate.
var monsterAnimationRuntimeNative533790 = func() server.MonsterAnimationNativeRuntime533790 {
	return server.MonsterAnimationNativeRuntime533790{
		Mimic:  func(unit *server.Object) uint32 { return uint32(C.nox_xxx_unitIsMimic_534840(asObjectC(unit))) },
		Plant:  func(unit *server.Object) uint32 { return uint32(C.nox_xxx_unitIsPlant_534A10(asObjectC(unit))) },
		Zombie: func(unit *server.Object) uint32 { return uint32(C.nox_xxx_unitIsZombie_534A40(asObjectC(unit))) },
	}
}
