package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

// HarpoonCollideData is the native-pointer form of HarpoonCollide's
// registered eight-byte ABI32 record. The callback itself does not read the
// record, but the HarpoonBolt producer stores its owner in the second field.
type HarpoonCollideData struct {
	Field0 uint32
	Owner  *Object
}

// HarpoonCollideRuntime4EB6A0 supplies state and effects that intentionally
// remain above package server. Native Object and PlayerUpdateData access stays
// inside HarpoonCollide4EB6A0.
type HarpoonCollideRuntime4EB6A0 struct {
	LoadDamage     func() int32
	StoreDamage    func(int32)
	DamageMap      func(int32, int32, int32, object.DamageType, *Object)
	DisableAbility func(*Object, Ability)
	DelayedDelete  func(*Object)
	MarkRelation   func(*Object, *Object)
}

type harpoonCollideNativeDeps4EB6A0 struct {
	loadDamage       func() int32
	loadBalance      func() float32
	floatToInt       func(float32) int32
	storeDamage      func(int32)
	damageMap        func(int32, int32, int32, object.DamageType, *Object)
	disableAbility   func(*Object, Ability)
	delayedDelete    func(*Object)
	markRelation     func(*Object, *Object)
	findParentPlayer func(*Object) *Object
	targetDamage     func(*Object, *Object, *Object, int32, object.DamageType) int32
	isEnemy          func(*Object, *Object) bool
	gameplayFlag     func(uint32) bool
	defaultSound     func(*Object, *Object)
	frame            func() uint32
	audio            func(uint32, *Object)
}

func harpoonCollideNative4EB6A0(
	source, target *Object,
	collision *types.Pointf,
	deps harpoonCollideNativeDeps4EB6A0,
) {
	harpoonCollide4EB6A0(source, target, collision, harpoonCollideHooks4EB6A0[
		*Object,
		*PlayerUpdateData,
	]{
		loadDamage: deps.loadDamage,
		loadOwner: func(obj *Object) *Object {
			return obj.ObjOwner
		},
		loadPlayerData: func(obj *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(obj.UpdateData)
		},
		loadBalanceDamage: deps.loadBalance,
		floatToInt:        deps.floatToInt,
		storeDamage:       deps.storeDamage,
		loadTargetFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		findParentPlayer: deps.findParentPlayer,
		targetDamage: func(target, parent, source *Object, damage int32, damageType uint32) int32 {
			return deps.targetDamage(target, parent, source, damage, object.DamageType(damageType))
		},
		isEnemy:      deps.isEnemy,
		gameplayFlag: deps.gameplayFlag,
		loadClassLo: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		loadNewPosY: func(obj *Object) float32 {
			return obj.NewPos.Y
		},
		loadNewPosX: func(obj *Object) float32 {
			return obj.NewPos.X
		},
		damageMap: func(x, y, damage int32, damageType uint32, source *Object) {
			deps.damageMap(x, y, damage, object.DamageType(damageType), source)
		},
		defaultDamageSound: deps.defaultSound,
		storeTarget: func(data *PlayerUpdateData, target *Object) {
			data.HarpoonTarg = target
		},
		disableAbility: func(owner *Object, ability int32) {
			deps.disableAbility(owner, Ability(ability))
		},
		delayedDelete: deps.delayedDelete,
		storeBolt: func(data *PlayerUpdateData, bolt *Object) {
			data.HarpoonBolt = bolt
		},
		loadPosX: func(obj *Object) float32 {
			return obj.PosVec.X
		},
		loadPosY: func(obj *Object) float32 {
			return obj.PosVec.Y
		},
		loadFrame: deps.frame,
		storeTargetX: func(data *PlayerUpdateData, value float32) {
			data.HarpoonTargX = value
		},
		storeTargetY: func(data *PlayerUpdateData, value float32) {
			data.HarpoonTargY = value
		},
		storeFrame: func(data *PlayerUpdateData, value uint32) {
			data.HarpoonFrame = value
		},
		loadSourceFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		storeSourceFlags: func(obj *Object, value uint32) {
			obj.ObjFlags = object.Flags(value)
		},
		markRelation: deps.markRelation,
		audio:        deps.audio,
	})
}

// harpoonRoundFloat32ToInt32_4EB6A0 models nox_float2int at 00419A70 under
// GAME.EXE's default x87 round-to-nearest-even mode. Invalid conversions
// produce the signed integer-indefinite value 0x80000000.
func harpoonRoundFloat32ToInt32_4EB6A0(value float32) int32 {
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int32(math.RoundToEven(float64(value)))
}

// HarpoonCollide4EB6A0 binds the registered collision callback to native-width
// Object, PlayerUpdateData and HarpoonCollideData layouts.
func (s *Server) HarpoonCollide4EB6A0(
	source, target *Object,
	collision *types.Pointf,
	runtime HarpoonCollideRuntime4EB6A0,
) {
	harpoonCollideNative4EB6A0(source, target, collision, harpoonCollideNativeDeps4EB6A0{
		loadDamage:       runtime.LoadDamage,
		loadBalance:      func() float32 { return float32(s.Balance.Float("HarpoonDamage")) },
		floatToInt:       harpoonRoundFloat32ToInt32_4EB6A0,
		storeDamage:      runtime.StoreDamage,
		damageMap:        runtime.DamageMap,
		disableAbility:   runtime.DisableAbility,
		delayedDelete:    runtime.DelayedDelete,
		markRelation:     runtime.MarkRelation,
		findParentPlayer: (*Object).FindOwnerChainPlayer,
		targetDamage: func(target, parent, source *Object, damage int32, damageType object.DamageType) int32 {
			return int32(ccall.CallIntUPtr5(
				target.Damage,
				uintptr(target.CObj()),
				uintptr(toObjectC(parent)),
				uintptr(source.CObj()),
				uintptr(uint32(damage)),
				uintptr(damageType),
			))
		},
		isEnemy: s.IsEnemyTo,
		gameplayFlag: func(flag uint32) bool {
			return noxflags.HasGamePlay(noxflags.GameplayFlag(flag))
		},
		defaultSound: func(target, source *Object) {
			Nox_xxx_soundDefaultDamageSound_532E20(target, source)
		},
		frame: s.Frame,
		audio: func(id uint32, owner *Object) {
			s.Audio.EventObj(sound.ID(id), owner, 0, 0)
		},
	})
}

var (
	_ = [1]struct{}{}[2*unsafe.Sizeof(uintptr(0))-unsafe.Sizeof(HarpoonCollideData{})]
	_ = [1]struct{}{}[unsafe.Sizeof(uintptr(0))-unsafe.Offsetof(HarpoonCollideData{}.Owner)]
)
