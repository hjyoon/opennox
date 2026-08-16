package legacy

/*
#include "GAME3_3.h"
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_spawnSomeBarrel_4E7470
func nox_xxx_spawnSomeBarrel_4E7470(source *nox_object_t, pos *C.float2) {
	srv := GetServer()
	s := srv.S()
	spawnSomeBarrel4E7470(asObjectS(source), barrelSpawnHooks4E7470[*server.Object, *server.Object, types.Pointf]{
		unitName: func(obj *server.Object) string {
			return s.Types.ByInd(int(obj.TypeInd)).ID()
		},
		randomInt: func(min, max int32) int32 {
			return int32(s.Rand.Logic.IntClamp(int(min), int(max)))
		},
		newObject: s.NewObjectByTypeID,
		randomPoint: func(radius float32) types.Pointf {
			// The source position is intentionally read only after allocation
			// succeeds, matching GAME.EXE's short-circuit behavior.
			var output types.Pointf
			s.RandomReachablePointAroundInto4ED970(
				radius,
				(*types.Pointf)(unsafe.Pointer(pos)),
				&output,
			)
			return output
		},
		createAt: func(obj, _ *server.Object, point types.Pointf) {
			srv.CreateObjectAt(obj, nil, point)
		},
	})
}
