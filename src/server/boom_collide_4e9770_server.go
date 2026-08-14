package server

import (
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

// boomCollideBalance4E9770 is the native, pointer-independent replacement for
// GAME.EXE globals 00753270..00753287.
type boomCollideBalance4E9770 struct {
	Ready        uint32
	DirectDamage int32
	SplashDamage int32
	Range        float32
	PushRange    float32
	Force        float32
}

// BoomCollideRuntime4E9770 supplies effects that remain owned by the legacy
// game runtime. Object fields, balance state, enemy checks, audio, networking,
// direction math and indirect damage dispatch remain native-width in server.
type BoomCollideRuntime4E9770 struct {
	CheckDirection  func(types.Pointf, int16, types.Pointf) int32
	ChangeOwner     func(*Object, *Object)
	Scorch          func(types.Pointf, int32)
	TraceHitPoint   func() *ntype.Point32
	DamageMap       func(int32, int32, int32, object.DamageType, *Object)
	MapDamageUnits  func(types.Pointf, float32, float32, int32, object.DamageType, *Object, *Object)
	MapPushUnits    func(types.Pointf, float32, float32, float32, *Object, int32, int32)
	DelayedDelete   func(*Object)
	InversionEffect unsafe.Pointer
}

type boomCollideNativeDeps4E9770 struct {
	balance        *boomCollideBalance4E9770
	gameDataFloat  func(string) float64
	floatToInt     func(float32) int32
	gameFlagsCheck func(uint32) int32
	findParent     func(*Object) *Object
	isEnemy        func(*Object, *Object) int32
	pointFX        func(uint32, types.Pointf)
	inversion      func(*Object, *Object) int32
	changeOwner    func(*Object, *Object)
	checkDirection func(types.Pointf, int16, types.Pointf) int32
	audio          func(uint32, *Object, int32, uint32)
	targetDamage   func(*Object, *Object, *Object, int32, object.DamageType) int32
	scorch         func(types.Pointf, int32)
	wallReflect    func(*types.Pointf, *Object)
	traceHitPoint  func() *ntype.Point32
	damageMap      func(int32, int32, int32, object.DamageType, *Object)
	mapDamageUnits func(types.Pointf, float32, float32, int32, object.DamageType, *Object, *Object)
	mapPushUnits   func(types.Pointf, float32, float32, float32, *Object, int32, int32)
	delayedDelete  func(*Object)
}

func boomCollideNative4E9770(
	source, target *Object,
	collision *types.Pointf,
	deps boomCollideNativeDeps4E9770,
) {
	boomCollide4E9770(source, target, collision, boomCollideHooks4E9770[
		*Object,
		*types.Pointf,
		*ntype.Point32,
	]{
		loadBalanceReady: func() uint32 {
			return deps.balance.Ready
		},
		gameDataFloat: deps.gameDataFloat,
		floatToInt:    deps.floatToInt,
		storeDirectDamage: func(value int32) {
			deps.balance.DirectDamage = value
		},
		storeSplashDamage: func(value int32) {
			deps.balance.SplashDamage = value
		},
		storeRange: func(value float32) {
			deps.balance.Range = value
		},
		storePushRange: func(value float32) {
			deps.balance.PushRange = value
		},
		storeForce: func(value float32) {
			deps.balance.Force = value
		},
		storeBalanceReady: func(value uint32) {
			deps.balance.Ready = value
		},
		gameFlagsCheck: deps.gameFlagsCheck,
		findParent:     deps.findParent,
		classLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		isEnemy: deps.isEnemy,
		pointFX: func(id uint32, obj *Object) {
			deps.pointFX(id, obj.PosVec)
		},
		inversion:   deps.inversion,
		changeOwner: deps.changeOwner,
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
		audio: deps.audio,
		loadDirectDamage: func() int32 {
			return deps.balance.DirectDamage
		},
		targetDamage: func(target, parent, source *Object, damage int32, damageType uint32) int32 {
			return deps.targetDamage(target, parent, source, damage, object.DamageType(damageType))
		},
		scorch: func(target *Object, kind int32) {
			deps.scorch(target.PosVec, kind)
		},
		wallReflect: deps.wallReflect,
		loadVelocityX: func(obj *Object) float32 {
			return obj.VelVec.X
		},
		loadVelocityY: func(obj *Object) float32 {
			return obj.VelVec.Y
		},
		vectorDirection: directionFromVector509ED0,
		storeDirection2: func(obj *Object, direction uint16) {
			obj.Direction2 = Dir16(direction)
		},
		storeVelocityX: func(obj *Object, value float32) {
			obj.VelVec.X = value
		},
		storeVelocityY: func(obj *Object, value float32) {
			obj.VelVec.Y = value
		},
		traceHitPoint: deps.traceHitPoint,
		loadPointY: func(point *ntype.Point32) int32 {
			return point.Y
		},
		loadPointX: func(point *ntype.Point32) int32 {
			return point.X
		},
		damageMap: func(x, y, damage int32, damageType uint32, source *Object) {
			deps.damageMap(x, y, damage, object.DamageType(damageType), source)
		},
		loadSplashDamage: func() int32 {
			return deps.balance.SplashDamage
		},
		loadRange: func() float32 {
			return deps.balance.Range
		},
		mapDamageUnits: func(pos *Object, radius, inner float32, damage int32, damageType uint32, source, excluded *Object) {
			deps.mapDamageUnits(pos.PosVec, radius, inner, damage, object.DamageType(damageType), source, excluded)
		},
		loadForce: func() float32 {
			return deps.balance.Force
		},
		loadPushRange: func() float32 {
			return deps.balance.PushRange
		},
		mapPushUnits: func(pos *Object, first, second, force float32, source *Object, arg6, arg7 int32) {
			deps.mapPushUnits(pos.PosVec, first, second, force, source, arg6, arg7)
		},
		delayedDelete: deps.delayedDelete,
	})
}

// BoomCollide4E9770 binds the BoomCollide callback and its zero-byte data
// record to native-width Object, Pointf and damage-callback boundaries.
func (s *Server) BoomCollide4E9770(
	source, target *Object,
	collision *types.Pointf,
	runtime BoomCollideRuntime4E9770,
) {
	boomCollideNative4E9770(source, target, collision, boomCollideNativeDeps4E9770{
		balance:       &s.boomBalance,
		gameDataFloat: s.Balance.Float,
		floatToInt:    playerCollideRound4E8460,
		gameFlagsCheck: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		findParent: (*Object).FindOwnerChainPlayer,
		isEnemy: func(first, second *Object) int32 {
			if s.IsEnemyTo(first, second) {
				return 1
			}
			return 0
		},
		pointFX: func(id uint32, pos types.Pointf) {
			s.Nox_xxx_netSendPointFx_522FF0(netmsg.Op(id), pos)
		},
		inversion: func(target, source *Object) int32 {
			return spellProjectileInversionNative4FA4F0(target, source, runtime.InversionEffect)
		},
		changeOwner:    runtime.ChangeOwner,
		checkDirection: runtime.CheckDirection,
		audio: func(id uint32, obj *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), obj, int(kind), code)
		},
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
		scorch:         runtime.Scorch,
		wallReflect:    spellProjectileWallReflect57B810,
		traceHitPoint:  runtime.TraceHitPoint,
		damageMap:      runtime.DamageMap,
		mapDamageUnits: runtime.MapDamageUnits,
		mapPushUnits:   runtime.MapPushUnits,
		delayedDelete:  runtime.DelayedDelete,
	})
}

var (
	_ = [1]struct{}{}[24-unsafe.Sizeof(boomCollideBalance4E9770{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(boomCollideBalance4E9770{}.Ready)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(boomCollideBalance4E9770{}.DirectDamage)]
	_ = [1]struct{}{}[8-unsafe.Offsetof(boomCollideBalance4E9770{}.SplashDamage)]
	_ = [1]struct{}{}[12-unsafe.Offsetof(boomCollideBalance4E9770{}.Range)]
	_ = [1]struct{}{}[16-unsafe.Offsetof(boomCollideBalance4E9770{}.PushRange)]
	_ = [1]struct{}{}[20-unsafe.Offsetof(boomCollideBalance4E9770{}.Force)]
)
