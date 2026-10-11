package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_playerAttackSelectCandidateNative_5386A0
func nox_xxx_playerAttackSelectCandidateNative_5386A0(candidate, attacker *nox_object_t, distance *C.float, closest **nox_object_t) {
	if candidate == nil || attacker == nil || distance == nil || closest == nil || candidate == attacker {
		return
	}
	s := GetServer().S()
	chosen := asObjectS(*closest)
	server.PlayerAttackSelectCandidate5386A0(asObjectS(candidate), asObjectS(attacker), (*float32)(unsafe.Pointer(distance)), &chosen,
		server.PlayerAttackTargetRuntime5386A0{
			IsEnemy: s.IsEnemyTo, CanInteract: s.CanInteract,
			Direction: func(direction server.Dir16) types.Pointf {
				// The original sign-extends the full WORD, not its low byte.
				offset := uintptr(int(int16(direction)) * 8)
				return types.Ptf(memmap.Float32(0x587000, 194136+offset), memmap.Float32(0x587000, 194140+offset))
			},
			Box: circleBoxCollisionNative54A990,
		})
	*closest = asObjectC(chosen)
}
