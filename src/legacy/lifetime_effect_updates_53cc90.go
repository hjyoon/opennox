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
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

// Construct world operations lazily so an unreached service cannot add a
// server lookup before the original required-pointer/early-return boundary.
var lifetimeEffectRuntime53CC90 = func() server.LifetimeEffectRuntime53CC90 {
	return server.LifetimeEffectRuntime53CC90{
		Frame: func() uint32 { return GetServer().S().Frame() }, FPS: func() uint32 { return GetServer().S().TickRate() },
		DelayedDelete: func(o *server.Object) { GetServer().DelayedDelete(o) }, BuffOff: Nox_xxx_spellBuffOff_4FF5B0,
		DamageAround: func(p *types.Pointf, a, b float32, d int, typ object.DamageType, o *server.Object, x server.Obj) {
			if x != nil {
				panic("fragment damage requires nil exclusion")
			}
			C.nox_xxx_mapDamageUnitsAround_4E25B0((*C.float)(unsafe.Pointer(p)), C.float(a), C.float(b), C.int(d), C.int(typ), asObjectC(o), nil)
		},
		ValidPosition: func(p types.Pointf) bool { return GetServer().S().Map.ValidIndexPos(p) },
		Move: func(o *server.Object, p *types.Pointf) {
			C.nox_xxx_unitMove_4E7010(asObjectC(o), (*C.float2)(unsafe.Pointer(p)))
		},
		SpellSound: func(id spell.ID, kind int) sound.ID {
			if kind != 2 {
				panic("effect sound requires off-sound selector")
			}
			return GetServer().S().Spells.DefByInd(id).GetOffSound()
		},
		Audio:    func(id sound.ID, o *server.Object) { GetServer().S().Audio.EventObj(id, o, 0, 0) },
		TraceRay: func(r *types.Rectf, f server.MapTraceFlags) bool { return GetServer().S().MapTraceRay(r.Min, r.Max, f) },
		Scorch:   func(p *types.Pointf, k int) { C.nox_xxx_sMakeScorch_537AF0((*C.float)(unsafe.Pointer(p)), C.int(k)) },
		PointFX: func(id uint8, p *types.Pointf) {
			C.nox_xxx_netSendPointFx_522FF0(C.char(id), (*C.float2)(unsafe.Pointer(p)))
		},
		Earthquake: func(p *types.Pointf, n int) { C.nox_xxx_earthquakeSend_4D9110((*C.float)(unsafe.Pointer(p)), C.int(n)) },
	}
}

//export nox_lifetime_effect_update_native
func nox_lifetime_effect_update_native(kind C.int, source *nox_object_t) {
	server.LifetimeEffectUpdateRaw53CC90(server.LifetimeEffectKind53CC90(kind), asObjectS(source), lifetimeEffectRuntime53CC90())
}
