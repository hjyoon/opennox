package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

// SparkExplosionCollideData is the one-byte record registered for
// SparkExplosionCollide in GAME.EXE.
type SparkExplosionCollideData struct {
	Power uint8
}

// SparkExplosionCollideRuntime4E9AC0 supplies effects that remain owned by
// the legacy runtime. Object fields, ownership, damage dispatch, audio and
// networking stay in the native-width server representation.
type SparkExplosionCollideRuntime4E9AC0 struct {
	CheckDirection func(types.Pointf, int16, types.Pointf) int32
	MapPushUnits   func(types.Pointf, float32, float32, float32, *Object, int32, int32)
	MapDamageUnits func(types.Pointf, float32, float32, int32, object.DamageType, *Object, *Object)
	Scorch         func(types.Pointf, int32)
	DelayedDelete  func(*Object)
}

type sparkExplosionCollideNativeDeps4E9AC0 struct {
	gameFlagsCheck func(uint32) int32
	checkDirection func(types.Pointf, int16, types.Pointf) int32
	reflect        func(*Object, *Object)
	clearOwner     func(*Object)
	setOwner       func(*Object, *Object)
	audio          func(uint32, *Object, int32, uint32)
	findParent     func(*Object) *Object
	isEnemy        func(*Object, *Object) int32
	mapPushUnits   func(types.Pointf, float32, float32, float32, *Object, int32, int32)
	targetDamage   func(*Object, *Object, *Object, int32, object.DamageType) int32
	mapDamageUnits func(types.Pointf, float32, float32, int32, object.DamageType, *Object, *Object)
	sparkFX        func(types.Pointf, uint8)
	scorch         func(types.Pointf, int32)
	delayedDelete  func(*Object)
}

func sparkExplosionCollideNative4E9AC0(
	source, target *Object,
	collision *types.Pointf,
	deps sparkExplosionCollideNativeDeps4E9AC0,
) {
	sparkExplosionCollide4E9AC0(
		source,
		target,
		collision,
		sparkExplosionCollideHooks4E9AC0[*Object, *SparkExplosionCollideData]{
			loadCollideData: func(obj *Object) *SparkExplosionCollideData {
				return (*SparkExplosionCollideData)(obj.CollideData)
			},
			loadPower: func(data *SparkExplosionCollideData) uint8 {
				return data.Power
			},
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
			reflect:        deps.reflect,
			clearOwner:     deps.clearOwner,
			setOwner:       deps.setOwner,
			audio:          deps.audio,
			gameFlagsCheck: deps.gameFlagsCheck,
			findParent:     deps.findParent,
			classLow: func(obj *Object) uint8 {
				return uint8(obj.ObjClass)
			},
			isEnemy: deps.isEnemy,
			mapPushUnits: func(pos *Object, first, second, force float32, owner *Object, arg6, arg7 int32) {
				deps.mapPushUnits(pos.PosVec, first, second, force, owner, arg6, arg7)
			},
			targetDamage: func(target, parent, source *Object, damage int32, damageType uint32) int32 {
				return deps.targetDamage(target, parent, source, damage, object.DamageType(damageType))
			},
			mapDamageUnits: func(pos *Object, radius, inner float32, damage int32, damageType uint32, source, excluded *Object) {
				deps.mapDamageUnits(pos.PosVec, radius, inner, damage, object.DamageType(damageType), source, excluded)
			},
			sparkFX: func(pos *Object, power uint8) {
				deps.sparkFX(pos.PosVec, power)
			},
			scorch: func(pos *Object, kind int32) {
				deps.scorch(pos.PosVec, kind)
			},
			delayedDelete: deps.delayedDelete,
		},
	)
}

// SparkExplosionCollide4E9AC0 binds the original callback's source, target,
// ignored collision pointer and one-byte data record to native-width fields.
func (s *Server) SparkExplosionCollide4E9AC0(
	source, target *Object,
	collision *types.Pointf,
	runtime SparkExplosionCollideRuntime4E9AC0,
) {
	sparkExplosionCollideNative4E9AC0(source, target, collision, sparkExplosionCollideNativeDeps4E9AC0{
		gameFlagsCheck: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		checkDirection: runtime.CheckDirection,
		reflect:        spellProjectileReflect4E0A70,
		clearOwner:     s.ObjClearOwner,
		setOwner:       s.ObjSetOwner,
		audio: func(id uint32, obj *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), obj, int(kind), code)
		},
		findParent: (*Object).FindOwnerChainPlayer,
		isEnemy: func(first, second *Object) int32 {
			if s.IsEnemyTo(first, second) {
				return 1
			}
			return 0
		},
		mapPushUnits: runtime.MapPushUnits,
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
		mapDamageUnits: runtime.MapDamageUnits,
		sparkFX:        s.Nox_xxx_netSparkExplosionFx_5231B0,
		scorch:         runtime.Scorch,
		delayedDelete:  runtime.DelayedDelete,
	})
}

var (
	_ = [1]struct{}{}[1-unsafe.Sizeof(SparkExplosionCollideData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(SparkExplosionCollideData{}.Power)]
)
