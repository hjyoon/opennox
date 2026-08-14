package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

// PixieCollideRuntime4EA080 supplies effects owned by the legacy-facing game
// runtime. Object and collide-data reads stay at the native-width server
// boundary.
type PixieCollideRuntime4EA080 struct {
	CheckDirection  func(types.Pointf, int16, types.Pointf) int32
	ChangeOwner     func(*Object, *Object)
	DamageMap       func(int32, int32, int32, object.DamageType, *Object)
	DelayedDelete   func(*Object)
	InversionEffect unsafe.Pointer
}

type pixieCollideNativeDeps4EA080 struct {
	isEnemy        func(*Object, *Object) int32
	checkInversion func(*Object, *Object) int32
	changeOwner    func(*Object, *Object)
	checkDirection func(types.Pointf, int16, types.Pointf) int32
	findParent     func(*Object) *Object
	targetDamage   func(*Object, *Object, *Object, int32, object.DamageType) int32
	audio          func(uint32, *Object)
	delayedDelete  func(*Object)
	wallReflect    func(*types.Pointf, *Object)
	floatToInt     func(float32) int32
	damageMap      func(int32, int32, int32, object.DamageType, *Object)
}

func pixieCollideNative4EA080(
	source, target *Object,
	collision *types.Pointf,
	deps pixieCollideNativeDeps4EA080,
) {
	pixieCollide4EA080(source, target, collision, pixieCollideHooks4EA080[
		*Object,
		*types.Pointf,
		*ProjectileCollideData,
	]{
		loadCollideData: func(obj *Object) *ProjectileCollideData {
			return (*ProjectileCollideData)(obj.CollideData)
		},
		isEnemy: deps.isEnemy,
		loadClass: func(obj *Object) uint32 {
			return uint32(obj.ObjClass)
		},
		loadFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		loadOwner: func(obj *Object) *Object {
			return obj.ObjOwner
		},
		checkInversion: deps.checkInversion,
		changeOwner:    deps.changeOwner,
		hasEnchant: func(obj *Object, enchant uint32) int32 {
			if obj.HasEnchant(EnchantID(enchant)) {
				return 1
			}
			return 0
		},
		loadDirection: func(obj *Object) int16 {
			return int16(obj.Direction1)
		},
		checkDirection: func(target *Object, direction int16, source *Object) int32 {
			return deps.checkDirection(target.PosVec, direction, source.PosVec)
		},
		loadDamage: func(data *ProjectileCollideData) int32 {
			return data.Damage
		},
		findParent: deps.findParent,
		targetDamage: func(target, parent, source *Object, damage int32, damageType uint32) int32 {
			return deps.targetDamage(target, parent, source, damage, object.DamageType(damageType))
		},
		audio:         deps.audio,
		delayedDelete: deps.delayedDelete,
		wallReflect:   deps.wallReflect,
		vectorDirection: func(obj *Object) int32 {
			return directionFromVector509ED0(obj.VelVec.X, obj.VelVec.Y)
		},
		loadVelocityX: func(obj *Object) float32 { return obj.VelVec.X },
		loadVelocityY: func(obj *Object) float32 { return obj.VelVec.Y },
		loadNewPosX:   func(obj *Object) float32 { return obj.NewPos.X },
		loadNewPosY:   func(obj *Object) float32 { return obj.NewPos.Y },
		storeDirection2: func(obj *Object, direction uint16) {
			obj.Direction2 = Dir16(direction)
		},
		storeNewPosX: func(obj *Object, value float32) { obj.NewPos.X = value },
		storeNewPosY: func(obj *Object, value float32) { obj.NewPos.Y = value },
		floatToInt:   deps.floatToInt,
		damageMap: func(x, y, damage int32, damageType uint32, source *Object) {
			deps.damageMap(x, y, damage, object.DamageType(damageType), source)
		},
	})
}

func pixieNativeDeps4EA080(
	s *Server,
	runtime PixieCollideRuntime4EA080,
) pixieCollideNativeDeps4EA080 {
	return pixieCollideNativeDeps4EA080{
		isEnemy: func(source, target *Object) int32 {
			if s.IsEnemyTo(source, target) {
				return 1
			}
			return 0
		},
		checkInversion: func(target, source *Object) int32 {
			return spellProjectileInversionNative4FA4F0(target, source, runtime.InversionEffect)
		},
		changeOwner:    runtime.ChangeOwner,
		checkDirection: runtime.CheckDirection,
		findParent:     (*Object).FindOwnerChainPlayer,
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
		audio: func(id uint32, obj *Object) {
			s.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
		delayedDelete: runtime.DelayedDelete,
		wallReflect:   spellProjectileWallReflect57B810,
		floatToInt:    playerCollideRound4E8460,
		damageMap:     runtime.DamageMap,
	}
}

// PixieCollide4EA080 binds the original callback to native Object and Pointf
// pointers while retaining fixed-width collide damage and map coordinates.
func (s *Server) PixieCollide4EA080(
	source, target *Object,
	collision *types.Pointf,
	runtime PixieCollideRuntime4EA080,
) {
	pixieCollideNative4EA080(source, target, collision, pixieNativeDeps4EA080(s, runtime))
}

var (
	_ = [1]struct{}{}[8-unsafe.Sizeof(ProjectileCollideData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(ProjectileCollideData{}.Damage)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(ProjectileCollideData{}.Field4)]
)
