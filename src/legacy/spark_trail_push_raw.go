package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

type sparkTrailPushRuntime struct {
	DelayedDelete func(*server.Object)
	Trail         server.ProjectileTrailRawRuntime53AEC0
	PushUnits     func(*types.Pointf, float32, float32, float32)
}

// Resolve world services only when the original routine reaches them.
var sparkTrailPushRawRuntime = func() sparkTrailPushRuntime {
	return sparkTrailPushRuntime{
		DelayedDelete: func(o *server.Object) { GetServer().DelayedDelete(o) },
		Trail: server.ProjectileTrailRawRuntime53AEC0{
			DirectionFloat: func(index int32) float32 { return *(*float32)(memmap.Ptr(uintptr(int64(0x5b6658) + int64(index)*4))) },
			RandomFloat:    func(lo, hi float32) float64 { return GetServer().S().RandomFloat416030(lo, hi) },
			CreateSpark: func(p types.Pointf, kind, life int, v types.Pointf, z float32, owner *server.Object) *server.Object {
				return Nox_xxx_createSpark_54FD80(p.X, p.Y, kind, life, v.X, v.Y, z, owner)
			},
		},
		PushUnits: func(p *types.Pointf, outer, inner, force float32) {
			GetServer().S().MapPushUnitsAroundPtr52E040(p, outer, inner, force, server.MapPushUnitsAroundRuntime52E040{
				Wake: Nox_xxx_unitHasCollideOrUpdateFn_537610,
			})
		},
	}
}

//export nox_spark_trail_push_native
func nox_spark_trail_push_native(kind C.int, source *nox_object_t) *nox_object_t {
	h := sparkTrailPushRawRuntime()
	switch kind {
	case 0:
		server.SparkUpdateRaw53ADC0(asObjectS(source), h.DelayedDelete)
	case 1:
		return asObjectC(server.ProjectileTrailUpdateRaw53AEC0(asObjectS(source), h.Trail))
	case 2:
		server.PushUpdateRaw53B030(asObjectS(source), h.PushUnits)
	default:
		panic("unknown spark/trail/push update")
	}
	return nil
}
