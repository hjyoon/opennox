package server

import (
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

func toxicCloudAffectTarget53D8C0(
	target, cloud *Object,
	randomDamage bool,
	deps toxicCloudUpdateDeps53D850,
) {
	if uint8(target.ObjClass)&0x6 == 0 {
		return
	}
	if !deps.traceRay(cloud.PosVec, target.PosVec, MapTraceFlags(5)) {
		return
	}
	damage := int32(0)
	if randomDamage {
		damage = int32(deps.randomInt(toxicCloudDamageMinimum53D8C0, toxicCloudDamageMaximum53D8C0))
	}
	owner := deps.findOwnerPlayer(cloud)
	deps.damage(target, owner, cloud, damage, object.DamagePoison)
	owner = deps.findOwnerPlayer(cloud)
	if deps.isEnemy(owner, target) {
		deps.activatePoison(target, 1, 1)
	}
}

func toxicCloudUpdateNative53D850(
	cloud *Object,
	radius float32,
	randomDamage bool,
	deps toxicCloudUpdateDeps53D850,
) {
	frame := deps.frame()
	nextEffectFrame := cloud.Field34
	data := (*ToxicCloudUpdateData)(cloud.UpdateData)
	if nextEffectFrame < frame {
		deps.eachInCircle(cloud.PosVec, radius, func(target *Object) bool {
			toxicCloudAffectTarget53D8C0(target, cloud, randomDamage, deps)
			return true
		})
		delay := deps.randomInt(toxicCloudDelayMinimum53D850, toxicCloudDelayMaximum53D850)
		cloud.Field34 = deps.frame() + uint32(delay)
	}
	if data.Duration != 0 {
		data.Duration--
	}
	if data.Duration == 0 {
		deps.delayedDelete(cloud)
	}
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
