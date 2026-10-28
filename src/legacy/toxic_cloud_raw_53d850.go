package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

// Preserve a raw indirect CALL's required callback. The general guarded
// damage registry intentionally accepts nil; these original C entries fault.
func toxicCloudCallDamageRaw53D8C0(fn unsafe.Pointer, target, owner, weapon *server.Object, d int32, typ object.DamageType) {
	if fn == nil {
		var required server.DamageFunc
		required(target, owner, weapon, d, typ)
	}
	server.CallObjectDamage(fn, target, owner, weapon, d, typ)
}

// Delay server lookup until a service is actually reached. A filtered target
// need not have a cloud/world, matching the original required-pointer gate.
var toxicCloudRawRuntime53D850 = func() server.ToxicCloudRawRuntime53D850 {
	return server.ToxicCloudRawRuntime53D850{
		Frame: func() uint32 { return GetServer().S().Frame() },
		EachInCircle: func(p *types.Pointf, r float32, visit func(*server.Object)) {
			GetServer().S().Map.EachObjInCircle(*p, r, func(o *server.Object) bool { visit(o); return true })
		},
		RandomInt: func(a, b int) int { return GetServer().S().Rand.Logic.IntClamp(a, b) },
		TraceRay: func(r *types.Rectf, flags server.MapTraceFlags) int32 {
			if GetServer().S().MapTraceRay(r.Min, r.Max, flags) {
				return 1
			}
			return 0
		},
		FindOwnerPlayer: (*server.Object).FindOwnerChainPlayer,
		Damage:          toxicCloudCallDamageRaw53D8C0,
		IsEnemy: func(a, b *server.Object) int32 {
			if GetServer().S().IsEnemyTo(a, b) {
				return 1
			}
			return 0
		},
		ActivatePoison: Nox_xxx_activatePoison_4EE7E0,
		DelayedDelete:  func(o *server.Object) { GetServer().DelayedDelete(o) },
	}
}

//export nox_toxic_cloud_root_native
func nox_toxic_cloud_root_native(small C.int, cloud *nox_object_t) {
	server.ToxicCloudUpdateRaw53D850(asObjectS(cloud), small != 0, toxicCloudRawRuntime53D850())
}

//export nox_toxic_cloud_target_native
func nox_toxic_cloud_target_native(small C.int, target, cloud *nox_object_t) {
	server.ToxicCloudAffectTargetRaw53D8C0(asObjectS(target), asObjectS(cloud), small != 0, toxicCloudRawRuntime53D850())
}
