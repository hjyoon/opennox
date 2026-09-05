package server

import (
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/sound"
)

type collisionEnchantNativeDeps4FDF90 struct {
	isEnemy           func(*Object, *Object) bool
	audio             func(int32, *Object, int32, uint32)
	disableEnchant    func(*Object, EnchantID)
	balanceFloatTable func(string, int32) float64
	floatToInt        func(float32) int32
	callDamage        func(unsafe.Pointer, *Object, *Object, *Object, int32, uint32) int32
	unitsOnSameTeam   func(*Object, *Object) bool
}

// CollisionEnchantRuntime4FDF90 supplies the enchant-removal side effect that
// intentionally remains above package server. Object reads, team predicates,
// audio, balance lookup, and damage dispatch stay on native-width values in
// CollisionEnchant4FDF90.
type CollisionEnchantRuntime4FDF90 struct {
	DisableEnchant func(*Object, EnchantID)
}

func collisionEnchantNative4FDF90(
	source, target *Object,
	deps collisionEnchantNativeDeps4FDF90,
) {
	collisionEnchant4FDF90(collisionEnchantHooks4FDF90[*Object]{
		loadSourceArg: func() *Object { return source },
		hasEnchant: func(obj *Object, enchant uint32) int32 {
			if obj.HasEnchant(EnchantID(enchant)) {
				return 1
			}
			return 0
		},
		loadTargetArg: func() *Object { return target },

		loadTargetFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		loadTargetClassLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		isEnemy: func(source, target *Object) int32 {
			if deps.isEnemy(source, target) {
				return 1
			}
			return 0
		},
		loadShockPower: func(obj *Object) uint8 {
			return obj.BuffsPower[ENCHANT_SHOCK]
		},
		audio: deps.audio,
		disableEnchant: func(obj *Object, enchant uint32) {
			deps.disableEnchant(obj, EnchantID(enchant))
		},
		balanceFloatTable: deps.balanceFloatTable,
		floatToInt:        deps.floatToInt,
		callTargetDamage: func(target, source, weapon *Object, damage int32, damageType uint32) int32 {
			// GAME.EXE loads the callback from target+716 immediately before
			// the indirect call. Keep this read live and deliberately do not
			// route through Object.CallDamage, which adds a nil guard.
			return deps.callDamage(target.Damage, target, source, weapon, damage, damageType)
		},

		loadTargetClass: func(obj *Object) uint32 {
			return uint32(obj.ObjClass)
		},
		unitsOnSameTeam: func(first, second *Object) int32 {
			if deps.unitsOnSameTeam(first, second) {
				return 1
			}
			return 0
		},
		loadSourceClassLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
	})
}

func collisionEnchantCallDamage4FDF90(
	fnc unsafe.Pointer,
	target, source, weapon *Object,
	damage int32,
	damageType uint32,
) int32 {
	// The original call through target+716 has no nil check. objDamage.Get(nil)
	// therefore intentionally yields a nil Go function and faults when called.
	if objDamage.Get(fnc)(target, source, weapon, damage, object.DamageType(damageType)) {
		return 1
	}
	return 0
}

// CollisionEnchant4FDF90 binds GAME.EXE 004FDF90 to native-width Object
// pointers while preserving every fixed-width class, flag, power, and damage
// operation of the original routine.
func (s *Server) CollisionEnchant4FDF90(
	source, target *Object,
	runtime CollisionEnchantRuntime4FDF90,
) {
	collisionEnchantNative4FDF90(source, target, collisionEnchantNativeDeps4FDF90{
		isEnemy: s.IsEnemyTo,
		audio: func(id int32, obj *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), obj, int(kind), code)
		},
		disableEnchant: runtime.DisableEnchant,
		balanceFloatTable: func(key string, index int32) float64 {
			return s.Balance.FloatInd(key, int(index))
		},
		floatToInt:      playerCollideRound4E8460,
		callDamage:      collisionEnchantCallDamage4FDF90,
		unitsOnSameTeam: UnitsHaveSameTeam4EC520,
	})
}
