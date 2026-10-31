package legacy

/*
#include "defs.h"
#include "GAME4.h"
#include "GAME4_3.h"
*/
import "C"

import (
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

// Construction only creates lazy callbacks. Idle NPC and nil-table paths
// must return without requesting a server or animation frame provider.
var monsterNPCAnimationRuntimeNative533D00 = func() server.MonsterNPCAnimationNativeRuntime533D00 {
	return server.MonsterNPCAnimationNativeRuntime533D00{
		WeaponAnimation: func(value uint32) int32 { return int32(C.sub_4FA280(C.int(value))) },
		Frames: func(index int32) (int32, int32) {
			var frames, duration C.int
			C.nox_xxx_animPlayerGetFrameRange_4F9F90(C.int(index), &frames, &duration)
			return int32(frames), int32(duration)
		},
		Selector: func(unit *server.Object) int32 {
			return int32(C.nox_xxx_mobActionToAnimation_533790(asObjectC(unit)))
		},
		Scratch: func() *server.MonsterAnim {
			return (*server.MonsterAnim)(memmap.PtrOff(0x5D4594, 2487964))
		},
	}
}
