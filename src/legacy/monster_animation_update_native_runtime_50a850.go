package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

// Keep the child lazy: completed, attack and nil-table paths need no server.
var monsterAnimationUpdateRuntimeNative50A850 = func() server.MonsterAnimationUpdateNativeRuntime50A850 {
	return server.MonsterAnimationUpdateNativeRuntime50A850{
		Animation: func(unit *server.Object) *server.MonsterAnim {
			return (*server.MonsterAnim)(unsafe.Pointer(C.nox_xxx_unitNPCActionToAnim_533D00(asObjectC(unit))))
		},
	}
}
