package legacy

/*
#include "defs.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4_2.h"
#include "GAME4_3.h"
*/
import "C"

import (
	"image"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

// Construct services lazily: waiting/airborne entries must not look up a
// world operation that the original body never reaches.
var meteorShowerRawRuntime53D5A0 = func() server.MeteorShowerRawRuntime53D5A0 {
	return server.MeteorShowerRawRuntime53D5A0{
		Frame: func() uint32 { return GetServer().S().Frame() }, FPS: func() uint32 { return GetServer().S().TickRate() },
		RandomInt:      func(a, b int) int { return GetServer().S().Rand.Logic.IntClamp(a, b) },
		RandomFloat:    func(a, b float64) float64 { return GetServer().S().RandomFloat416030(float32(a), float32(b)) },
		DirectionFloat: func(index int32) float32 { return *(*float32)(memmap.Ptr(uintptr(int64(0x5b6658) + int64(index)*4))) },
		TraceRay: func(r *types.Rectf, _ *server.Object) int32 {
			return int32(C.nox_xxx_traceRay_5374B0((*C.float4)(unsafe.Pointer(r))))
		},
		NewObjectByTypeID: func(name string) *server.Object { return GetServer().S().NewObjectByTypeID(name) },
		CreateAt: func(o, owner *server.Object, p types.Pointf) {
			C.nox_xxx_createAt_4DAA50(asObjectC(o), asObjectC(owner), C.float(p.X), C.float(p.Y))
		},
		Raise:         func(o *server.Object, z float32) { C.nox_xxx_unitRaise_4E46F0(asObjectC(o), C.float(z)) },
		DelayedDelete: func(o *server.Object) { GetServer().DelayedDelete(o) },
	}
}

var meteorRawRuntime53D6E0 = func() server.MeteorRawRuntime53D6E0 {
	return server.MeteorRawRuntime53D6E0{
		AudioEvent:        func(id uint32, o *server.Object) { GetServer().S().Audio.EventObj(sound.ID(id), o, 0, 0) },
		MakeScorch:        func(p *types.Pointf, k int) { C.nox_xxx_sMakeScorch_537AF0((*C.float)(unsafe.Pointer(p)), C.int(k)) },
		Earthquake:        func(p *types.Pointf, k int) { C.nox_xxx_earthquakeSend_4D9110((*C.float)(unsafe.Pointer(p)), C.int(k)) },
		NewObjectByTypeID: func(name string) *server.Object { return GetServer().S().NewObjectByTypeID(name) },
		CreateAt: func(o, owner *server.Object, p types.Pointf) {
			C.nox_xxx_createAt_4DAA50(asObjectC(o), asObjectC(owner), C.float(p.X), C.float(p.Y))
		},
		FindParentChainPlayer: (*server.Object).FindOwnerChainPlayer,
		DamageUnitsAround: func(p *types.Pointf, a, b float32, d int, typ object.DamageType, o *server.Object, x server.Obj) {
			if x != nil {
				panic("meteor damage requires nil exclusion")
			}
			C.nox_xxx_mapDamageUnitsAround_4E25B0((*C.float)(unsafe.Pointer(p)), C.float(a), C.float(b), C.int(d), C.int(typ), asObjectC(o), nil)
		},
		DamageWalls: func(r image.Rectangle, p *types.Pointf, a float32, d int, typ object.DamageType, o *server.Object) {
			// Preserve the original signed bounds; the older C export's
			// image.Rect conversion would reorder a reversed raw interval.
			GetServer().Nox_xxx_mapDamageToWalls_534FC0(r, *p, a, d, typ, o)
		},
		DelayedDelete: func(o *server.Object) { GetServer().DelayedDelete(o) },
	}
}

//export nox_meteor_update_native
func nox_meteor_update_native(kind C.int, source *nox_object_t) {
	switch kind {
	case 0:
		server.MeteorShowerUpdateRaw53D5A0(asObjectS(source), meteorShowerRawRuntime53D5A0())
	case 1:
		server.MeteorUpdateRaw53D6E0(asObjectS(source), meteorRawRuntime53D6E0())
	default:
		panic("unknown meteor update")
	}
}
