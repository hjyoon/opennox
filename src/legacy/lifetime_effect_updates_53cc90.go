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
		Earthquake: sendEarthquakeFromNative,
	}
}

func updateLifetimeEffect(kind server.LifetimeEffectKind53CC90, source *nox_object_t) {
	server.LifetimeEffectUpdateRaw53CC90(kind, asObjectS(source), lifetimeEffectRuntime53CC90())
}

//export nox_xxx_updateSelfDestruct_53CC90
func nox_xxx_updateSelfDestruct_53CC90(source *nox_object_t) {
	updateLifetimeEffect(server.LifetimeEffectSelfDestruct53CC90, source)
}

//export nox_xxx_updateDeathBallFragment_53D220
func nox_xxx_updateDeathBallFragment_53D220(source *nox_object_t) {
	updateLifetimeEffect(server.LifetimeEffectDeathBallFragment53D220, source)
}

//export nox_xxx_updateMoonglow_53D270
func nox_xxx_updateMoonglow_53D270(source *nox_object_t) {
	updateLifetimeEffect(server.LifetimeEffectMoonglow53D270, source)
}

//export nox_xxx_updateTelekinesis_53D330
func nox_xxx_updateTelekinesis_53D330(source *nox_object_t) {
	updateLifetimeEffect(server.LifetimeEffectTelekinesis53D330, source)
}

//export nox_xxx_updateFist_53D400
func nox_xxx_updateFist_53D400(source *nox_object_t) {
	updateLifetimeEffect(server.LifetimeEffectFist53D400, source)
}

//export nox_xxx_updateFlameCleanse_53D510
func nox_xxx_updateFlameCleanse_53D510(source *nox_object_t) {
	updateLifetimeEffect(server.LifetimeEffectFlameCleanse53D510, source)
}
