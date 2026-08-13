package server

import (
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

// ProjectileCollideData is the pointer-independent eight-byte collide record
// allocated for ProjectileCollide and ProjectileSparkCollide object types.
// GAME.EXE 004E87B0 and 004E8880 read only Damage; the second word is retained
// because it is part of the registered record width.
type ProjectileCollideData struct {
	Damage int32
	Field4 int32
}

type projectileCollideNativeDeps4E87B0 struct {
	loadThrowingStoneType func() uint32
	lookupType            func(string) uint32
	storeThrowingStone    func(uint32)
	storeImpShot          func(uint32)
	loadImpShotType       func() uint32
	gameDataFloat         func(string) float64
	floatToInt            func(float32) int32
	traceHitPoint         func() *ntype.Point32
	damageMap             func(int32, int32, int32, object.DamageType, *Object)
	delayedDelete         func(*Object)
}

// ProjectileCollideRuntime4E87B0 supplies the map-trace and object lifecycle
// effects owned by the legacy runtime. Native object fields and callback
// dispatch remain inside ProjectileCollide4E87B0.
type ProjectileCollideRuntime4E87B0 struct {
	TraceHitPoint func() *ntype.Point32
	DamageMap     func(int32, int32, int32, object.DamageType, *Object)
	DelayedDelete func(*Object)
}

func projectileCollideNative4E87B0(
	projectile, other *Object,
	collision unsafe.Pointer,
	deps projectileCollideNativeDeps4E87B0,
) {
	projectileCollide4E87B0(projectile, other, collision, projectileCollideHooks4E87B0[
		*Object,
		unsafe.Pointer,
		*ntype.Point32,
		*ProjectileCollideData,
	]{
		loadCollideData: func(obj *Object) *ProjectileCollideData {
			return (*ProjectileCollideData)(obj.CollideData)
		},
		loadThrowingStoneType: deps.loadThrowingStoneType,
		lookupType:            deps.lookupType,
		storeThrowingStone:    deps.storeThrowingStone,
		storeImpShot:          deps.storeImpShot,
		loadType: func(obj *Object) uint16 {
			return obj.TypeInd
		},
		loadImpShotType: deps.loadImpShotType,
		gameDataFloat:   deps.gameDataFloat,
		floatToInt:      deps.floatToInt,
		loadDamage: func(data *ProjectileCollideData) int32 {
			return data.Damage
		},
		findParentPlayer: (*Object).FindOwnerChainPlayer,
		damage: func(target, source, attacker *Object, damage int32, damageType uint32) uint8 {
			return uint8(ccall.CallIntUPtr5(
				target.Damage,
				uintptr(target.CObj()),
				uintptr(toObjectC(source)),
				uintptr(toObjectC(attacker)),
				uintptr(uint32(damage)),
				uintptr(damageType),
			))
		},
		traceHitPoint: deps.traceHitPoint,
		loadPointY: func(point *ntype.Point32) int32 {
			return point.Y
		},
		loadPointX: func(point *ntype.Point32) int32 {
			return point.X
		},
		damageMap: func(x, y, damage int32, damageType uint32, projectile *Object) {
			deps.damageMap(x, y, damage, object.DamageType(damageType), projectile)
		},
		delayedDelete: deps.delayedDelete,
	})
}

// ProjectileCollide4E87B0 binds the original generic projectile collision to
// native-pointer Object and ProjectileCollideData records.
func (s *Server) ProjectileCollide4E87B0(
	projectile, other *Object,
	collision unsafe.Pointer,
	runtime ProjectileCollideRuntime4E87B0,
) {
	projectileCollideNative4E87B0(projectile, other, collision, projectileCollideNativeDeps4E87B0{
		loadThrowingStoneType: func() uint32 {
			return s.Types.fast.throwingStone
		},
		lookupType: func(name string) uint32 {
			return uint32(s.Types.IndByID(name))
		},
		storeThrowingStone: func(value uint32) {
			s.Types.fast.throwingStone = value
		},
		storeImpShot: func(value uint32) {
			s.Types.fast.impShot = value
		},
		loadImpShotType: func() uint32 {
			return s.Types.fast.impShot
		},
		gameDataFloat: s.Balance.Float,
		floatToInt:    playerCollideRound4E8460,
		traceHitPoint: runtime.TraceHitPoint,
		damageMap:     runtime.DamageMap,
		delayedDelete: runtime.DelayedDelete,
	})
}

var (
	_ = [1]struct{}{}[8-unsafe.Sizeof(ProjectileCollideData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(ProjectileCollideData{}.Damage)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(ProjectileCollideData{}.Field4)]
)
