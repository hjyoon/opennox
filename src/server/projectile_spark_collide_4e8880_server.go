package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
)

type projectileSparkCollideNativeDeps4E8880 struct {
	floatToInt    func(float32) int32
	damageMap     func(int32, int32, int32, object.DamageType, *Object)
	delayedDelete func(*Object)
}

// ProjectileSparkCollideRuntime4E8880 supplies the map and object lifecycle
// effects owned by the legacy runtime. Native object reads and damage callback
// dispatch remain inside ProjectileSparkCollide4E8880.
type ProjectileSparkCollideRuntime4E8880 struct {
	DamageMap     func(int32, int32, int32, object.DamageType, *Object)
	DelayedDelete func(*Object)
}

func projectileSparkCollideNative4E8880(
	projectile, other *Object,
	collision unsafe.Pointer,
	deps projectileSparkCollideNativeDeps4E8880,
) {
	projectileSparkCollide4E8880(projectile, other, collision, projectileSparkCollideHooks4E8880[
		*Object,
		*ProjectileCollideData,
	]{
		loadCollideData: func(obj *Object) *ProjectileCollideData {
			return (*ProjectileCollideData)(obj.CollideData)
		},
		loadDamage: func(data *ProjectileCollideData) int32 {
			return data.Damage
		},
		findParent: (*Object).FindOwnerChainPlayer,
		damage: func(target, source, attacker *Object, damage int32, damageType uint32) uint8 {
			return uint8(callObjectDamageNativeResult(
				target.Damage,
				target,
				source,
				attacker,
				damage,
				object.DamageType(damageType),
			))
		},
		loadNewPosY: func(obj *Object) float32 { return obj.NewPos.Y },
		loadNewPosX: func(obj *Object) float32 { return obj.NewPos.X },
		floatToInt:  deps.floatToInt,
		damageMap: func(x, y, damage int32, damageType uint32, projectile *Object) {
			deps.damageMap(x, y, damage, object.DamageType(damageType), projectile)
		},
		delayedDelete: deps.delayedDelete,
	})
}

// ProjectileSparkCollide4E8880 binds the original projectile spark collision
// to native-pointer Object and ProjectileCollideData records.
func (s *Server) ProjectileSparkCollide4E8880(
	projectile, other *Object,
	collision unsafe.Pointer,
	runtime ProjectileSparkCollideRuntime4E8880,
) {
	projectileSparkCollideNative4E8880(projectile, other, collision, projectileSparkCollideNativeDeps4E8880{
		floatToInt:    playerCollideRound4E8460,
		damageMap:     runtime.DamageMap,
		delayedDelete: runtime.DelayedDelete,
	})
}
