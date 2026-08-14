package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

// DeathBallFragmentCollideRuntime4E9FE0 supplies the legacy-owned map and
// lifecycle effects. Object fields and the indirect Damage callback remain at
// the native-width server boundary.
type DeathBallFragmentCollideRuntime4E9FE0 struct {
	DamageMap     func(int32, int32, int32, object.DamageType, *Object)
	DelayedDelete func(*Object)
}

type deathBallFragmentCollideNativeDeps4E9FE0 struct {
	findParent    func(*Object) *Object
	targetDamage  func(*Object, *Object, *Object, int32, object.DamageType) int32
	wallReflect   func(*types.Pointf, *Object)
	audio         func(uint32, *Object)
	floatToInt    func(float32) int32
	damageMap     func(int32, int32, int32, object.DamageType, *Object)
	delayedDelete func(*Object)
}

func deathBallFragmentCollideNative4E9FE0(
	source, target *Object,
	collision *types.Pointf,
	deps deathBallFragmentCollideNativeDeps4E9FE0,
) {
	deathBallFragmentCollide4E9FE0(
		source,
		target,
		collision,
		deathBallFragmentCollideHooks4E9FE0[*Object, *types.Pointf]{
			findParent: deps.findParent,
			targetDamage: func(target, parent, source *Object, damage int32, damageType uint32) int32 {
				return deps.targetDamage(target, parent, source, damage, object.DamageType(damageType))
			},
			wallReflect: deps.wallReflect,
			audio:       deps.audio,
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
			delayedDelete: deps.delayedDelete,
		},
	)
}

func deathBallFragmentNativeDeps4E9FE0(
	s *Server,
	runtime DeathBallFragmentCollideRuntime4E9FE0,
) deathBallFragmentCollideNativeDeps4E9FE0 {
	return deathBallFragmentCollideNativeDeps4E9FE0{
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
		wallReflect: spellProjectileWallReflect57B810,
		audio: func(id uint32, obj *Object) {
			s.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
		floatToInt:    playerCollideRound4E8460,
		damageMap:     runtime.DamageMap,
		delayedDelete: runtime.DelayedDelete,
	}
}

// DeathBallFragmentCollide4E9FE0 binds the original callback to native Object
// and Pointf pointers while keeping map coordinates and damage fixed-width.
func (s *Server) DeathBallFragmentCollide4E9FE0(
	source, target *Object,
	collision *types.Pointf,
	runtime DeathBallFragmentCollideRuntime4E9FE0,
) {
	deathBallFragmentCollideNative4E9FE0(
		source,
		target,
		collision,
		deathBallFragmentNativeDeps4E9FE0(s, runtime),
	)
}
