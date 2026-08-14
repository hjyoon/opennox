package server

import (
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

// WallReflectCollideRuntime4E9D80 supplies the legacy-owned callback identity,
// map effect, and object lifecycle operation. Object and collide-data reads,
// team/owner traversal, wall reflection, and indirect damage dispatch stay in
// the native-width server adapter.
type WallReflectCollideRuntime4E9D80 struct {
	YellowStarCollide unsafe.Pointer
	DamageMap         func(int32, int32, int32, object.DamageType, *Object)
	DelayedDelete     func(*Object)
}

type wallReflectCollideNativeDeps4E9D80 struct {
	sameTeam          func(*Object, *Object) int32
	gameFlagsCheck    func(uint32) int32
	yellowStarCollide unsafe.Pointer
	findParent        func(*Object) *Object
	targetDamage      func(*Object, *Object, *Object, int32, object.DamageType) int32
	delayedDelete     func(*Object)
	wallReflect       func(*types.Pointf, *Object)
	floatToInt        func(float32) int32
	damageMap         func(int32, int32, int32, object.DamageType, *Object)
}

func wallReflectCollideNative4E9D80(
	source, target *Object,
	collision *types.Pointf,
	deps wallReflectCollideNativeDeps4E9D80,
) {
	wallReflectCollide4E9D80(source, target, collision, wallReflectCollideHooks4E9D80[
		*Object,
		*types.Pointf,
		unsafe.Pointer,
		*ProjectileCollideData,
	]{
		loadCollideData: func(obj *Object) *ProjectileCollideData {
			return (*ProjectileCollideData)(obj.CollideData)
		},
		sameTeam:       deps.sameTeam,
		gameFlagsCheck: deps.gameFlagsCheck,
		loadCollide: func(obj *Object) unsafe.Pointer {
			return obj.Collide
		},
		yellowStarCollide: deps.yellowStarCollide,
		loadDamage: func(data *ProjectileCollideData) int32 {
			return data.Damage
		},
		findParent: deps.findParent,
		targetDamage: func(target, parent, source *Object, damage int32, damageType uint32) int32 {
			return deps.targetDamage(target, parent, source, damage, object.DamageType(damageType))
		},
		delayedDelete: deps.delayedDelete,
		wallReflect:   deps.wallReflect,
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

type yellowStarShotCollideNativeDeps4E9E50 struct {
	gameFlagsCheck func(uint32) int32
	pointFX        func(uint32, types.Pointf)
	wallCollide    func(*Object, *Object, *types.Pointf)
}

func yellowStarShotCollideNative4E9E50(
	source, target *Object,
	collision *types.Pointf,
	deps yellowStarShotCollideNativeDeps4E9E50,
) {
	yellowStarShotCollide4E9E50(source, target, collision, yellowStarShotCollideHooks4E9E50[
		*Object,
		*types.Pointf,
	]{
		gameFlagsCheck: deps.gameFlagsCheck,
		pointFX: func(id uint32, obj *Object) {
			deps.pointFX(id, obj.PosVec)
		},
		wallCollide: deps.wallCollide,
	})
}

func wallReflectNativeDeps4E9D80(
	s *Server,
	runtime WallReflectCollideRuntime4E9D80,
) wallReflectCollideNativeDeps4E9D80 {
	return wallReflectCollideNativeDeps4E9D80{
		sameTeam: func(first, second *Object) int32 {
			if UnitsHaveSameTeam4EC520(first, second) {
				return 1
			}
			return 0
		},
		gameFlagsCheck: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		yellowStarCollide: runtime.YellowStarCollide,
		findParent:        (*Object).FindOwnerChainPlayer,
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
		wallReflect:   spellProjectileWallReflect57B810,
		floatToInt:    playerCollideRound4E8460,
		damageMap:     runtime.DamageMap,
	}
}

// WallReflectCollide4E9D80 binds the original generic wall-reflect callback
// to native Object pointers and the fixed-width eight-byte projectile record.
func (s *Server) WallReflectCollide4E9D80(
	source, target *Object,
	collision *types.Pointf,
	runtime WallReflectCollideRuntime4E9D80,
) {
	wallReflectCollideNative4E9D80(source, target, collision, wallReflectNativeDeps4E9D80(s, runtime))
}

// YellowStarShotCollide4E9E50 emits the original point FX before forwarding
// to WallReflectCollide4E9D80 with the same native pointers.
func (s *Server) YellowStarShotCollide4E9E50(
	source, target *Object,
	collision *types.Pointf,
	runtime WallReflectCollideRuntime4E9D80,
) {
	wallDeps := wallReflectNativeDeps4E9D80(s, runtime)
	yellowStarShotCollideNative4E9E50(source, target, collision, yellowStarShotCollideNativeDeps4E9E50{
		gameFlagsCheck: wallDeps.gameFlagsCheck,
		pointFX: func(id uint32, pos types.Pointf) {
			s.Nox_xxx_netSendPointFx_522FF0(netmsg.Op(id), pos)
		},
		wallCollide: func(source, target *Object, collision *types.Pointf) {
			wallReflectCollideNative4E9D80(source, target, collision, wallDeps)
		},
	})
}

var (
	_ = [1]struct{}{}[8-unsafe.Sizeof(ProjectileCollideData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(ProjectileCollideData{}.Damage)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(ProjectileCollideData{}.Field4)]
)
