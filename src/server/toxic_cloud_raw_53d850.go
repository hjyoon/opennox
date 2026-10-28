package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// ToxicCloudRawRuntime53D850 retains native point, object and damage-function
// identities at the original world service boundaries.
type ToxicCloudRawRuntime53D850 struct {
	Frame           func() uint32
	EachInCircle    func(*types.Pointf, float32, func(*Object))
	RandomInt       func(int, int) int
	TraceRay        func(*types.Rectf, MapTraceFlags) int32
	FindOwnerPlayer func(*Object) *Object
	Damage          func(unsafe.Pointer, *Object, *Object, *Object, int32, object.DamageType)
	IsEnemy         func(*Object, *Object) int32
	ActivatePoison  func(*Object, int32, int32) int32
	DelayedDelete   func(*Object)
}

// ToxicCloudUpdateRaw53D850 preserves the required-pointer C root contracts.
func ToxicCloudUpdateRaw53D850(cloud *Object, small bool, h ToxicCloudRawRuntime53D850) {
	radius := toxicCloudRadius53D850
	if small {
		radius = smallToxicCloudRadius53D960
	}
	toxicCloudUpdateRaw53D850(cloud, radius, small, h)
}

func toxicCloudUpdateRaw53D850(cloud *Object, radius float32, small bool, h ToxicCloudRawRuntime53D850) {
	frame := h.Frame()
	next := cloud.Field34
	data := (*ToxicCloudUpdateData)(cloud.UpdateData)
	if next < frame {
		h.EachInCircle(&cloud.PosVec, radius, func(target *Object) {
			ToxicCloudAffectTargetRaw53D8C0(target, cloud, small, h)
		})
		delay := uint32(h.RandomInt(toxicCloudDelayMinimum53D850, toxicCloudDelayMaximum53D850))
		cloud.Field34 = h.Frame() + delay
	}
	if data.Duration != 0 {
		data.Duration--
	}
	if data.Duration == 0 {
		h.DelayedDelete(cloud)
	}
}

// ToxicCloudAffectTargetRaw53D8C0 preserves the early class gate and the live
// callback load between owner lookup and damage. Damage return values do not
// suppress the second owner lookup, enemy test or poison operation.
func ToxicCloudAffectTargetRaw53D8C0(target, cloud *Object, small bool, h ToxicCloudRawRuntime53D850) {
	if uint8(target.ObjClass)&6 == 0 {
		return
	}
	ray := types.Rectf{Min: cloud.PosVec, Max: target.PosVec}
	if h.TraceRay(&ray, MapTraceFlags(5)) == 0 {
		return
	}
	damage := int32(0)
	if !small {
		damage = int32(h.RandomInt(toxicCloudDamageMinimum53D8C0, toxicCloudDamageMaximum53D8C0))
	}
	owner := h.FindOwnerPlayer(cloud)
	callback := target.Damage
	h.Damage(callback, target, owner, cloud, damage, object.DamagePoison)
	owner = h.FindOwnerPlayer(cloud)
	if h.IsEnemy(owner, target) != 0 {
		h.ActivatePoison(target, 1, 1)
	}
}
