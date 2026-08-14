package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

// DeathBallCollideRuntime4E9E90 supplies the legacy trace globals and map
// damage callback. Object, Door, balance, audio, direction-table, and indirect
// Damage operations stay in the native-width server adapter.
type DeathBallCollideRuntime4E9E90 struct {
	TraceReady func() uint32
	TracePoint func() *ntype.Point32
	DamageMap  func(int32, int32, int32, object.DamageType, *Object)
}

type deathBallCollideNativeDeps4E9E90 struct {
	balanceFloat  func(string) float64
	floatToInt    func(float32) int32
	findParent    func(*Object) *Object
	targetDamage  func(*Object, *Object, *Object, int32, object.DamageType) int32
	doorReflect   func(*Object, float32, float32)
	wallReflect   func(*types.Pointf, *Object)
	audio         func(uint32, *Object)
	traceHitPoint func() *ntype.Point32
	damageMap     func(int32, int32, int32, object.DamageType, *Object)
}

func deathBallCollideNative4E9E90(
	source, target *Object,
	collision *types.Pointf,
	deps deathBallCollideNativeDeps4E9E90,
) {
	deathBallCollide4E9E90(source, target, collision, deathBallCollideHooks4E9E90[
		*Object,
		*types.Pointf,
		*DoorUpdateData,
		*ntype.Point32,
	]{
		loadClassByte: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		loadDoorUpdate: (*Object).UpdateDataDoor,
		loadPrevX: func(obj *Object) float32 {
			return obj.PrevPos.X
		},
		loadPrevY: func(obj *Object) float32 {
			return obj.PrevPos.Y
		},
		storeNewX: func(obj *Object, value float32) {
			obj.NewPos.X = value
		},
		storeNewY: func(obj *Object, value float32) {
			obj.NewPos.Y = value
		},
		loadPosX: func(obj *Object) float32 {
			return obj.PosVec.X
		},
		loadPosY: func(obj *Object) float32 {
			return obj.PosVec.Y
		},
		loadDoorDirection: func(update *DoorUpdateData) int32 {
			return update.CurrentDirection
		},
		loadDirectionY: func(direction int32) int32 {
			return deathBallDoorDirection4E9E90(direction).y
		},
		loadDirectionX: func(direction int32) int32 {
			return deathBallDoorDirection4E9E90(direction).x
		},
		doorReflect:  deps.doorReflect,
		audio:        deps.audio,
		balanceFloat: deps.balanceFloat,
		floatToInt:   deps.floatToInt,
		findParent:   deps.findParent,
		targetDamage: func(target, parent, source *Object, damage int32, damageType uint32) int32 {
			return deps.targetDamage(target, parent, source, damage, object.DamageType(damageType))
		},
		wallReflect:   deps.wallReflect,
		traceHitPoint: deps.traceHitPoint,
		loadTraceY: func(point *ntype.Point32) int32 {
			return point.Y
		},
		loadTraceX: func(point *ntype.Point32) int32 {
			return point.X
		},
		damageMap: func(x, y, damage int32, damageType uint32, source *Object) {
			deps.damageMap(x, y, damage, object.DamageType(damageType), source)
		},
	})
}

func deathBallNativeDeps4E9E90(
	s *Server,
	runtime DeathBallCollideRuntime4E9E90,
) deathBallCollideNativeDeps4E9E90 {
	return deathBallCollideNativeDeps4E9E90{
		balanceFloat: s.Balance.Float,
		floatToInt:   playerCollideRound4E8460,
		findParent:   (*Object).FindOwnerChainPlayer,
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
		doorReflect: func(source *Object, normalX, normalY float32) {
			deathBallDoorReflectCore57B770(
				source,
				normalX,
				normalY,
				deathBallDoorReflectHooks57B770[*Object]{
					loadVelocityX: func(obj *Object) float32 { return obj.VelVec.X },
					loadVelocityY: func(obj *Object) float32 { return obj.VelVec.Y },
					storeVelocityX: func(obj *Object, value float32) {
						obj.VelVec.X = value
					},
					storeVelocityY: func(obj *Object, value float32) {
						obj.VelVec.Y = value
					},
				},
			)
		},
		wallReflect: spellProjectileWallReflect57B810,
		audio: func(id uint32, obj *Object) {
			s.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
		traceHitPoint: func() *ntype.Point32 {
			return deathBallTraceHitResult537760(runtime.TraceReady, runtime.TracePoint)
		},
		damageMap: runtime.DamageMap,
	}
}

// DeathBallCollide4E9E90 binds the original callback to native Object and
// Pointf pointers while retaining fixed-width trace coordinates and damage.
func (s *Server) DeathBallCollide4E9E90(
	source, target *Object,
	collision *types.Pointf,
	runtime DeathBallCollideRuntime4E9E90,
) {
	deathBallCollideNative4E9E90(source, target, collision, deathBallNativeDeps4E9E90(s, runtime))
}

var (
	_ = [1]struct{}{}[8-unsafe.Sizeof(ntype.Point32{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(ntype.Point32{}.X)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(ntype.Point32{}.Y)]
)
