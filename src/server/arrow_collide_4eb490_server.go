package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

// ArrowCollideData is the native-pointer form of ArrowCollide's registered
// eight-byte ABI32 record. GAME.EXE leaves the first word unidentified and
// stores the firing Object pointer at +4. Natural pointer alignment moves
// Owner to +8 and widens the record to 16 bytes on 64-bit targets.
type ArrowCollideData struct {
	Field0 uint32
	Owner  *Object
}

// ArrowAttackData is the native-pointer form of GAME.EXE's initialized
// 32-byte stack attack record. Only its two Object references widen.
type ArrowAttackData = arrowAttack4EB490[*Object]

// ArrowCollideRuntime4EB490 lists effects that still cross the legacy-facing
// runtime. Object fields, modifier lookup, game rules, type caching, damage
// dispatch and conversion remain native-width in server.
type ArrowCollideRuntime4EB490 struct {
	TraceHitPoint     func() *ntype.Point32
	DamageMap         func(int32, int32, int32, object.DamageType, *Object)
	DelayedDelete     func(*Object)
	ApplyAttackEffect func(*Object, *Object, *ArrowAttackData)
	PreAttackEffects  func(*Object, *Object, *Object, *ArrowAttackData)
}

type arrowCollideNativeDeps4EB490 struct {
	lookupProjectileClass func(uint16) *Modifier
	strength              func(*Object) int32
	gameFlag              func(uint32) bool
	findParentPlayer      func(*Object) *Object
	isEnemy               func(*Object, *Object) bool
	tracePoint            func() (int32, int32, bool)
	calcBoltDamage        func(int32, *Modifier) float64
	floatToInt            func(float64) int32
	damageMap             func(int32, int32, int32, uint32, *Object)
	delayedDelete         func(*Object)
	loadArcherBoltType    func() uint32
	lookupType            func(string) uint32
	storeArcherBoltType   func(uint32)
	applyAttackEffect     func(*Object, *Object, *ArrowAttackData)
	preAttackEffects      func(*Object, *Object, *Object, *ArrowAttackData)
	targetDamage          func(*Object, *Object, *Object, int32, uint32) int32
}

func arrowCollideNative4EB490(
	source, target *Object,
	collision *types.Pointf,
	deps arrowCollideNativeDeps4EB490,
) {
	arrowCollide4EB490(source, target, collision, arrowCollideHooks4EB490[
		*Object,
		*ArrowCollideData,
		*Modifier,
		*HealthData,
	]{
		loadTypeIndex: func(obj *Object) uint16 {
			return obj.TypeInd
		},
		loadCollideData: func(obj *Object) *ArrowCollideData {
			return (*ArrowCollideData)(obj.CollideData)
		},
		lookupProjectileClass: deps.lookupProjectileClass,
		loadOwner: func(obj *Object) *Object {
			return obj.ObjOwner
		},
		strength:         deps.strength,
		gameFlag:         deps.gameFlag,
		findParentPlayer: deps.findParentPlayer,
		loadClassLo: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		isEnemy:             deps.isEnemy,
		tracePoint:          deps.tracePoint,
		calcBoltDamage:      deps.calcBoltDamage,
		floatToInt:          deps.floatToInt,
		damageMap:           deps.damageMap,
		delayedDelete:       deps.delayedDelete,
		loadArcherBoltType:  deps.loadArcherBoltType,
		lookupType:          deps.lookupType,
		storeArcherBoltType: deps.storeArcherBoltType,
		loadFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
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
		loadDataOwner: func(data *ArrowCollideData) *Object {
			return data.Owner
		},
		applyAttackEffect: deps.applyAttackEffect,
		preAttackEffects:  deps.preAttackEffects,
		targetDamage:      deps.targetDamage,
		loadHealth: func(obj *Object) *HealthData {
			return obj.HealthData
		},
		loadHealthCur: func(health *HealthData) uint16 {
			return health.Cur
		},
		loadHealthMax: func(health *HealthData) uint16 {
			return health.Max
		},
	})
}

// arrowTruncFloat64ToInt32_4EB490 models GAME.EXE's 00566DCC helper: truncate
// to a signed 64-bit integer with x87 FISTP, then return its low 32 bits.
// Invalid qword conversions produce 0x8000000000000000, whose low word is 0.
func arrowTruncFloat64ToInt32_4EB490(value float64) int32 {
	if math.IsNaN(value) || value >= 0x1p63 || value < -0x1p63 {
		return 0
	}
	return int32(int64(value))
}

func arrowCollideServerDeps4EB490(
	s *Server,
	runtime ArrowCollideRuntime4EB490,
) arrowCollideNativeDeps4EB490 {
	return arrowCollideNativeDeps4EB490{
		lookupProjectileClass: func(index uint16) *Modifier {
			return s.Modif.Nox_xxx_getProjectileClassById413250(int(index))
		},
		strength: func(obj *Object) int32 {
			return int32(obj.Strength())
		},
		gameFlag: func(flag uint32) bool {
			return noxflags.HasGame(noxflags.GameFlag(flag))
		},
		findParentPlayer: (*Object).FindOwnerChainPlayer,
		isEnemy:          s.IsEnemyTo,
		tracePoint: func() (int32, int32, bool) {
			point := runtime.TraceHitPoint()
			if point == nil {
				return 0, 0, false
			}
			return point.X, point.Y, true
		},
		calcBoltDamage: func(strength int32, modifier *Modifier) float64 {
			return chakramCalcBoltDamage4EF1E0(
				strength,
				modifier,
				noxflags.HasGame(noxflags.GameModeCoop),
				uint32(s.Types.IndByID(chakramArcherBoltTypeName4EF1E0)),
				s.Balance.Float("BoltSoloDamageMin"),
			)
		},
		floatToInt: arrowTruncFloat64ToInt32_4EB490,
		damageMap: func(x, y, damage int32, damageType uint32, source *Object) {
			runtime.DamageMap(x, y, damage, object.DamageType(damageType), source)
		},
		delayedDelete: runtime.DelayedDelete,
		loadArcherBoltType: func() uint32 {
			return s.Types.fast.arrowBolt
		},
		lookupType: func(name string) uint32 {
			return uint32(s.Types.IndByID(name))
		},
		storeArcherBoltType: func(value uint32) {
			s.Types.fast.arrowBolt = value
		},
		applyAttackEffect: runtime.ApplyAttackEffect,
		preAttackEffects:  runtime.PreAttackEffects,
		targetDamage: func(target, parent, source *Object, damage int32, damageType uint32) int32 {
			return int32(ccall.CallIntUPtr5(
				target.Damage,
				uintptr(target.CObj()),
				uintptr(toObjectC(parent)),
				uintptr(source.CObj()),
				uintptr(uint32(damage)),
				uintptr(damageType),
			))
		},
	}
}

// ArrowCollide4EB490 binds the registered collision callback to native-width
// Object, Pointf, collide-data, health and attack records.
func (s *Server) ArrowCollide4EB490(
	source, target *Object,
	collision *types.Pointf,
	runtime ArrowCollideRuntime4EB490,
) {
	arrowCollideNative4EB490(source, target, collision, arrowCollideServerDeps4EB490(s, runtime))
}

var (
	_ = [1]struct{}{}[2*unsafe.Sizeof(uintptr(0))-unsafe.Sizeof(ArrowCollideData{})]
	_ = [1]struct{}{}[unsafe.Sizeof(uintptr(0))-unsafe.Offsetof(ArrowCollideData{}.Owner)]
	_ = [1]struct{}{}[16+4*unsafe.Sizeof(uintptr(0))-unsafe.Sizeof(ArrowAttackData{})]
	_ = [1]struct{}{}[8+unsafe.Sizeof(uintptr(0))-unsafe.Offsetof(ArrowAttackData{}.Owner)]
	_ = [1]struct{}{}[16+3*unsafe.Sizeof(uintptr(0))-unsafe.Offsetof(ArrowAttackData{}.Source)]
)
