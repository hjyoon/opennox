package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

// WallReflectSparkCollideRuntime4EA200 supplies the map and lifecycle effects
// owned by the legacy-facing runtime. Object, collision, and collide-data
// access remains at this native-width server boundary.
type WallReflectSparkCollideRuntime4EA200 struct {
	DamageMap     func(int32, int32, int32, object.DamageType, *Object)
	DelayedDelete func(*Object)
}

type wallReflectSparkCollideNativeDeps4EA200 struct {
	findParent    func(*Object) *Object
	targetDamage  func(*Object, *Object, *Object, int32, object.DamageType) int32
	delayedDelete func(*Object)
	floatToInt    func(float32) int32
	damageMap     func(int32, int32, int32, object.DamageType, *Object)
}

func wallReflectSparkCollideNative4EA200(
	source, target *Object,
	collision *types.Pointf,
	deps wallReflectSparkCollideNativeDeps4EA200,
) {
	wallReflectSparkCollide4EA200(source, target, collision, wallReflectSparkCollideHooks4EA200[
		*Object,
		*types.Pointf,
		*ProjectileCollideData,
	]{
		loadCollideData: func(obj *Object) *ProjectileCollideData {
			return (*ProjectileCollideData)(obj.CollideData)
		},
		loadDamage: func(data *ProjectileCollideData) int32 {
			return data.Damage
		},
		findParent: deps.findParent,
		targetDamage: func(target, parent, source *Object, damage int32, damageType uint32) int32 {
			return deps.targetDamage(target, parent, source, damage, object.DamageType(damageType))
		},
		delayedDelete: deps.delayedDelete,
		loadCollisionY: func(point *types.Pointf) float32 {
			return point.Y
		},
		loadCollisionX: func(point *types.Pointf) float32 {
			return point.X
		},
		loadVelocityX: func(obj *Object) float32 {
			return obj.VelVec.X
		},
		loadVelocityY: func(obj *Object) float32 {
			return obj.VelVec.Y
		},
		storeVelocityX: func(obj *Object, value float32) {
			obj.VelVec.X = value
		},
		storeVelocityY: func(obj *Object, value float32) {
			obj.VelVec.Y = value
		},
		loadNewPosY: func(obj *Object) float32 {
			return obj.NewPos.Y
		},
		loadNewPosX: func(obj *Object) float32 {
			return obj.NewPos.X
		},
		floatToInt: deps.floatToInt,
		damageMap: func(x, y, damage int32, damageType uint32, source *Object) {
			deps.damageMap(x, y, damage, object.DamageType(damageType), source)
		},
	})
}

func wallReflectSparkNativeDeps4EA200(
	runtime WallReflectSparkCollideRuntime4EA200,
) wallReflectSparkCollideNativeDeps4EA200 {
	return wallReflectSparkCollideNativeDeps4EA200{
		findParent: (*Object).FindOwnerChainPlayer,
		targetDamage: func(target, parent, source *Object, damage int32, damageType object.DamageType) int32 {
			return int32(ccall.CallIntUPtr5(
				target.Damage,
				uintptr(target.CObj()),
				uintptr(toObjectC(parent)),
				uintptr(toObjectC(source)),
				uintptr(uint32(damage)),
				uintptr(uint32(damageType)),
			))
		},
		delayedDelete: runtime.DelayedDelete,
		floatToInt:    playerCollideRound4E8460,
		damageMap:     runtime.DamageMap,
	}
}

// WallReflectSparkCollide4EA200 binds the original callback to native Object
// and Pointf pointers while retaining its fixed-width projectile data.
func (s *Server) WallReflectSparkCollide4EA200(
	source, target *Object,
	collision *types.Pointf,
	runtime WallReflectSparkCollideRuntime4EA200,
) {
	wallReflectSparkCollideNative4EA200(source, target, collision, wallReflectSparkNativeDeps4EA200(runtime))
}

var (
	_ = [1]struct{}{}[8-unsafe.Sizeof(ProjectileCollideData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(ProjectileCollideData{}.Damage)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(ProjectileCollideData{}.Field4)]
)
