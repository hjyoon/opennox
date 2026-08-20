package server

import (
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

// ChakramAttackData is the native-pointer form of the initialized 32-byte
// attack record built on the stack by GAME.EXE. It remains 32 bytes on a
// 32-bit target and widens only its Object pointers on a 64-bit target.
type ChakramAttackData = chakramAttack4EAF00[*Object]

// ChakramUpdateData is the native-pointer form of the original 28-byte
// ChakramInMotionUpdate record. Field0 and the three bytes following
// Reflections have not yet been assigned domain names by recovered code.
type ChakramUpdateData struct {
	Field0       uint32
	Reflections  uint8
	_            [3]byte
	ReturnTarget *Object
	LastHit      *Object
	OwnerPos     types.Pointf
	ReturnState  uint8
}

// ChakramCollideRuntime4EAF00 lists the remaining effects owned by the
// legacy-facing runtime. Object fields, math, team checks, modifier lookup,
// damage dispatch, networking and audio stay native-width in server.
type ChakramCollideRuntime4EAF00 struct {
	TraceHitPoint     func() *ntype.Point32
	DamageMap         func(int32, int32, int32, object.DamageType, *Object)
	Drop              func(*Object, *Object, *types.Pointf)
	DelayedDelete     func(*Object)
	MoveUpdate        func(*Object)
	DetachInventory   func(*Object, *Object)
	InventoryPut      func(*Object, *Object, uint32)
	EquipWeapon       func(*Object, *Object, uint32, uint32)
	ApplyAttackEffect func(*Object, *Object, *ChakramAttackData)
	PreAttackEffects  func(*Object, *Object, *Object, *ChakramAttackData)
	CreateAt          func(*Object, *Object, types.Pointf)
}

type chakramCollideNativeDeps4EAF00 struct {
	ownerHasWeapon        func(*Object) bool
	pointFX               func(uint32, *types.Pointf)
	wallReflect           func(*types.Pointf, *types.Pointf)
	randomReflect         func(*Object)
	tracePoint            func() (int32, int32, bool)
	damageMap             func(int32, int32, int32, uint32, *Object)
	drop                  func(*Object, *Object, *types.Pointf)
	delayedDelete         func(*Object)
	retarget              func(*Object)
	detach                func(*Object, *Object)
	inventoryPut          func(*Object, *Object, uint32)
	equipWeapon           func(*Object, *Object, uint32, uint32)
	audio                 func(uint32, *Object)
	sameTeam              func(*Object, *Object) bool
	lookupProjectileClass func(uint16) *Modifier
	strength              func(*Object) int32
	calcBoltDamage        func(int32, *Modifier) float32
	applyAttackEffect     func(*Object, *Object, *ChakramAttackData)
	preAttackEffects      func(*Object, *Object, *Object, *ChakramAttackData)
	floatToInt            func(float64) int32
	targetDamage          func(*Object, *Object, *Object, int32, uint32)
	projectileReflect     func(*Object, *Object)
	createAt              func(*Object, *Object, *types.Pointf)
}

// chakramCollideNative4EAF00 binds the recovered control flow to native-width
// Object, Modifier, Pointf, attack and update records. Operations whose owning
// subsystems are still being restored are explicit dependencies rather than
// ABI32 integer-pointer calls.
func chakramCollideNative4EAF00(
	source, target *Object,
	collision *types.Pointf,
	deps chakramCollideNativeDeps4EAF00,
) {
	chakramCollide4EAF00(source, target, collision, chakramCollideHooks4EAF00[
		*Object,
		*ChakramUpdateData,
		*types.Pointf,
		*types.Pointf,
		*types.Pointf,
		*Modifier,
	]{
		loadUpdateData: func(obj *Object) *ChakramUpdateData {
			return (*ChakramUpdateData)(obj.UpdateData)
		},
		inventoryFirst: func(obj *Object) *Object {
			return obj.InvFirstItem
		},
		loadFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		loadMaterialLo: func(obj *Object) uint8 {
			return uint8(obj.Material)
		},
		loadOwner: func(obj *Object) *Object {
			return obj.ObjOwner
		},
		loadClassLo: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		ownerHasWeapon: deps.ownerHasWeapon,
		loadTypeIndex: func(obj *Object) uint16 {
			return obj.TypeInd
		},
		loadPosX: func(obj *Object) float32 {
			return obj.PosVec.X
		},
		loadPosY: func(obj *Object) float32 {
			return obj.PosVec.Y
		},
		loadRadius: func(obj *Object) float32 {
			return obj.Shape.Circle.R
		},
		position: func(obj *Object) *types.Pointf {
			return &obj.PosVec
		},
		velocity: func(obj *Object) *types.Pointf {
			return &obj.VelVec
		},
		loadReflections: func(update *ChakramUpdateData) uint8 {
			return update.Reflections
		},
		storeReflections: func(update *ChakramUpdateData, value uint8) {
			update.Reflections = value
		},
		loadReturnState: func(update *ChakramUpdateData) uint8 {
			return update.ReturnState
		},
		storeReturnState: func(update *ChakramUpdateData, value uint8) {
			update.ReturnState = value
		},
		storeReturnTarget: func(update *ChakramUpdateData, obj *Object) {
			update.ReturnTarget = obj
		},
		storeLastHit: func(update *ChakramUpdateData, obj *Object) {
			update.LastHit = obj
		},
		pointFX:               deps.pointFX,
		wallReflect:           deps.wallReflect,
		randomReflect:         deps.randomReflect,
		tracePoint:            deps.tracePoint,
		damageMap:             deps.damageMap,
		drop:                  deps.drop,
		delayedDelete:         deps.delayedDelete,
		retarget:              deps.retarget,
		detach:                deps.detach,
		inventoryPut:          deps.inventoryPut,
		equipWeapon:           deps.equipWeapon,
		audio:                 deps.audio,
		sameTeam:              deps.sameTeam,
		lookupProjectileClass: deps.lookupProjectileClass,
		strength:              deps.strength,
		calcBoltDamage:        deps.calcBoltDamage,
		applyAttackEffect:     deps.applyAttackEffect,
		preAttackEffects:      deps.preAttackEffects,
		floatToInt:            deps.floatToInt,
		targetDamage:          deps.targetDamage,
		projectileReflect:     deps.projectileReflect,
		createAt:              deps.createAt,
	})
}

func chakramCollideServerDeps4EAF00(
	s *Server,
	runtime ChakramCollideRuntime4EAF00,
) chakramCollideNativeDeps4EAF00 {
	return chakramCollideNativeDeps4EAF00{
		ownerHasWeapon: func(owner *Object) bool {
			return (*PlayerUpdateData)(owner.UpdateData).EquippedWeapon != nil
		},
		pointFX: func(id uint32, pos *types.Pointf) {
			s.Nox_xxx_netSendPointFx_522FF0(netmsg.Op(id), *pos)
		},
		wallReflect: func(collision, velocity *types.Pointf) {
			x, y := velocity.X, velocity.Y
			if !(float64(collision.Y)*float64(collision.X) > 0) {
				velocity.X = y
				velocity.Y = x
				return
			}
			velocity.X = -y
			velocity.Y = -x
		},
		randomReflect: func(source *Object) {
			chakramRandomReflectNative4EB3E0(source, chakramRandomReflectNativeDeps4EB3E0{
				randomInt: func(minimum, maximum int32) int32 {
					return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
				},
				moveUpdate: runtime.MoveUpdate,
			})
		},
		tracePoint: func() (int32, int32, bool) {
			point := runtime.TraceHitPoint()
			if point == nil {
				return 0, 0, false
			}
			return point.X, point.Y, true
		},
		damageMap: func(x, y, damage int32, damageType uint32, source *Object) {
			runtime.DamageMap(x, y, damage, object.DamageType(damageType), source)
		},
		drop: func(owner, item *Object, pos *types.Pointf) {
			runtime.Drop(owner, item, pos)
		},
		delayedDelete: runtime.DelayedDelete,
		retarget: func(source *Object) {
			chakramRetargetNative4EB250(source, chakramRetargetNativeDeps4EB250{
				eachInRect: func(rect types.Rectf, callback func(*Object)) {
					s.Map.EachObjInRect(rect, func(obj *Object) bool {
						callback(obj)
						return true
					})
				},
				mapCheck: s.MapTraceVision,
			})
		},
		detach:       runtime.DetachInventory,
		inventoryPut: runtime.InventoryPut,
		equipWeapon:  runtime.EquipWeapon,
		audio: func(id uint32, obj *Object) {
			s.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
		sameTeam: UnitsHaveSameTeam4EC520,
		lookupProjectileClass: func(index uint16) *Modifier {
			return s.Modif.Nox_xxx_getProjectileClassById413250(int(index))
		},
		strength: func(obj *Object) int32 {
			return int32(obj.Strength())
		},
		calcBoltDamage: func(strength int32, modifier *Modifier) float32 {
			return float32(s.CalcBoltDamage4EF1E0(strength, modifier))
		},
		applyAttackEffect: runtime.ApplyAttackEffect,
		preAttackEffects:  runtime.PreAttackEffects,
		floatToInt: func(value float64) int32 {
			return int32(value)
		},
		targetDamage: func(target, owner, source *Object, damage int32, damageType uint32) {
			ccall.CallIntUPtr5(
				target.Damage,
				uintptr(target.CObj()),
				uintptr(toObjectC(owner)),
				uintptr(toObjectC(source)),
				uintptr(uint32(damage)),
				uintptr(damageType),
			)
		},
		projectileReflect: spellProjectileReflect4E0A70,
		createAt: func(item, owner *Object, pos *types.Pointf) {
			runtime.CreateAt(item, owner, *pos)
		},
	}
}

// ChakramInMotionCollide4EAF00 binds the registered collision callback to
// native-width Object, Pointf, update-data and attack-data records.
func (s *Server) ChakramInMotionCollide4EAF00(
	source, target *Object,
	collision *types.Pointf,
	runtime ChakramCollideRuntime4EAF00,
) {
	chakramCollideNative4EAF00(source, target, collision, chakramCollideServerDeps4EAF00(s, runtime))
}

var (
	_ = [1]struct{}{}[0-unsafe.Offsetof(ChakramUpdateData{}.Field0)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(ChakramUpdateData{}.Reflections)]
	_ = [1]struct{}{}[8-unsafe.Offsetof(ChakramUpdateData{}.ReturnTarget)]
	_ = [1]struct{}{}[0-unsafe.Offsetof(ChakramAttackData{}.Damage)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(ChakramAttackData{}.DamageType)]
	_ = [1]struct{}{}[8-unsafe.Offsetof(ChakramAttackData{}.Radius)]
)
