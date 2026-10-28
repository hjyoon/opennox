package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

const (
	toxicCloudRadius53D850        = float32(75)
	smallToxicCloudRadius53D960   = float32(35)
	toxicCloudDamageMinimum53D8C0 = 3
	toxicCloudDamageMaximum53D8C0 = 10
	toxicCloudDelayMinimum53D850  = 5
	toxicCloudDelayMaximum53D850  = 10
)

// ToxicCloudUpdateRuntime53D850 supplies effects that remain owned by the
// legacy-facing runtime. The map, random stream, object callbacks, and owner
// traversal are bound by Server without sending object pointers through C.
type ToxicCloudUpdateRuntime53D850 struct {
	ActivatePoison func(*Object, int32, int32) int32
	DelayedDelete  func(*Object)
}

type toxicCloudUpdateDeps53D850 struct {
	frame           func() uint32
	eachInCircle    func(types.Pointf, float32, func(*Object) bool)
	randomInt       func(int, int) int
	traceRay        func(types.Pointf, types.Pointf, MapTraceFlags) bool
	findOwnerPlayer func(*Object) *Object
	damage          func(*Object, *Object, *Object, int32, object.DamageType)
	isEnemy         func(*Object, *Object) bool
	activatePoison  func(*Object, int32, int32) int32
	delayedDelete   func(*Object)
}

func toxicCloudNativeRuntime53D850(deps toxicCloudUpdateDeps53D850) ToxicCloudRawRuntime53D850 {
	return ToxicCloudRawRuntime53D850{
		Frame: deps.frame, RandomInt: deps.randomInt,
		EachInCircle: func(p *types.Pointf, r float32, visit func(*Object)) {
			deps.eachInCircle(*p, r, func(o *Object) bool { visit(o); return true })
		},
		TraceRay: func(r *types.Rectf, flags MapTraceFlags) int32 {
			if deps.traceRay(r.Min, r.Max, flags) {
				return 1
			}
			return 0
		},
		FindOwnerPlayer: deps.findOwnerPlayer,
		Damage:          func(_ unsafe.Pointer, t, s, w *Object, d int32, typ object.DamageType) { deps.damage(t, s, w, d, typ) },
		IsEnemy: func(a, b *Object) int32 {
			if deps.isEnemy(a, b) {
				return 1
			}
			return 0
		},
		ActivatePoison: deps.activatePoison, DelayedDelete: deps.delayedDelete,
	}
}

func toxicCloudAffectTarget53D8C0(target, cloud *Object, randomDamage bool, deps toxicCloudUpdateDeps53D850) {
	ToxicCloudAffectTargetRaw53D8C0(target, cloud, !randomDamage, toxicCloudNativeRuntime53D850(deps))
}

func toxicCloudUpdateNative53D850(cloud *Object, radius float32, randomDamage bool, deps toxicCloudUpdateDeps53D850) {
	toxicCloudUpdateRaw53D850(cloud, radius, !randomDamage, toxicCloudNativeRuntime53D850(deps))
}

func (s *Server) toxicCloudUpdateDeps53D850(runtime ToxicCloudUpdateRuntime53D850) toxicCloudUpdateDeps53D850 {
	return toxicCloudUpdateDeps53D850{
		frame:        s.Frame,
		eachInCircle: s.Map.EachObjInCircle,
		randomInt:    s.Rand.Logic.IntClamp,
		traceRay:     s.MapTraceRay,
		findOwnerPlayer: func(obj *Object) *Object {
			return obj.FindOwnerChainPlayer()
		},
		damage: func(target, source, weapon *Object, damage int32, damageType object.DamageType) {
			target.CallDamage(source, weapon, int(damage), damageType)
		},
		isEnemy:        s.IsEnemyTo,
		activatePoison: runtime.ActivatePoison,
		delayedDelete:  runtime.DelayedDelete,
	}
}

// ToxicCloudUpdate53D850 restores GAME.EXE 0053D850 and its 0053D8C0
// target callback with native-width Object and owner pointers.
func (s *Server) ToxicCloudUpdate53D850(cloud *Object, runtime ToxicCloudUpdateRuntime53D850) {
	if cloud == nil || cloud.UpdateData == nil {
		return
	}
	toxicCloudUpdateNative53D850(
		cloud,
		toxicCloudRadius53D850,
		true,
		s.toxicCloudUpdateDeps53D850(runtime),
	)
}

// SmallToxicCloudUpdate53D960 restores GAME.EXE 0053D960 and its 0053D9D0
// zero-damage target callback with native-width Object and owner pointers.
func (s *Server) SmallToxicCloudUpdate53D960(cloud *Object, runtime ToxicCloudUpdateRuntime53D850) {
	if cloud == nil || cloud.UpdateData == nil {
		return
	}
	toxicCloudUpdateNative53D850(
		cloud,
		smallToxicCloudRadius53D960,
		false,
		s.toxicCloudUpdateDeps53D850(runtime),
	)
}
