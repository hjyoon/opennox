package server

import (
	"math"

	"github.com/opennox/libs/object"
)

const playerDamageWeaponExcludedSubclass4E1560 = uint32(0x07800000)

// ItemDurabilityDamageRuntime4E1560 contains the callbacks reached by the
// weapon and equipped-item durability routines. ApplyDefend receives the
// original argument order: modifier, item, owner, effective weapon, source.
type ItemDurabilityDamageRuntime4E1560 struct {
	ApplyDefend  func(*ModifierEff, *Object, *Object, *Object, *Object, *float32) bool
	Damage       func(*Object, *Object, *Object, int32, object.DamageType) bool
	ReportHealth func(*Object, *Object, uint16, uint16)
	Unsupported  func(string, *Object, *Object, *Object, *Object, float32, object.DamageType)
}

func itemDurabilityUnsupported4E1560(
	runtime ItemDurabilityDamageRuntime4E1560,
	reason string,
	item, owner, source, effective *Object,
	amount float32,
	typ object.DamageType,
) bool {
	if runtime.Unsupported != nil {
		runtime.Unsupported(reason, item, owner, source, effective, amount, typ)
	}
	return false
}

// itemDurabilityRound4E1560 models the x87 FISTP conversion reached through
// nox_float2int. The original returns the integer-indefinite value for NaN and
// values outside the signed 32-bit range.
func itemDurabilityRound4E1560(value float32) int32 {
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int32(math.RoundToEven(float64(value)))
}

func itemDurabilityDamage4E1560(
	item, owner, source, effective *Object,
	amount float32,
	typ object.DamageType,
	weaponRules bool,
	runtime ItemDurabilityDamageRuntime4E1560,
) bool {
	if item == nil || item.HealthData == nil {
		return true
	}
	if weaponRules && item.ObjClass.Has(object.ClassWeapon) &&
		uint32(item.ObjSubClass)&playerDamageWeaponExcludedSubclass4E1560 != 0 {
		return true
	}
	if item.UpdateData == nil {
		return itemDurabilityUnsupported4E1560(
			runtime, "nil durability update data", item, owner, source, effective, amount, typ,
		)
	}
	if item.InitData == nil {
		return itemDurabilityUnsupported4E1560(
			runtime, "nil modifier init data", item, owner, source, effective, amount, typ,
		)
	}

	modifier := item.InitDataModifier().Modifiers[1]
	if modifier != nil && modifier.Defend76.Fnc != nil {
		if runtime.ApplyDefend == nil ||
			!runtime.ApplyDefend(modifier, item, owner, effective, source, &amount) {
			return itemDurabilityUnsupported4E1560(
				runtime, "unsupported durability modifier", item, owner, source, effective, amount, typ,
			)
		}
	}

	update := item.UpdateDataWeaponArmor()
	total := amount + math.Float32frombits(update.Field0)
	damage := itemDurabilityRound4E1560(total)
	if damage > 0 && (item.Damage == nil || runtime.Damage == nil) {
		return itemDurabilityUnsupported4E1560(
			runtime, "missing item damage callback", item, owner, source, effective, amount, typ,
		)
	}

	update.Field0 = math.Float32bits(total - float32(damage))
	if damage <= 0 {
		return true
	}

	before := item.HealthData.Cur
	runtime.Damage(item, source, effective, damage, typ)
	if item.HealthData == nil {
		return true
	}
	after := item.HealthData.Cur
	if before != after && owner != nil && owner.ObjClass.Has(object.ClassPlayer) && runtime.ReportHealth != nil {
		runtime.ReportHealth(owner, item, before, after)
	}
	return true
}

// PlayerDamageWeaponNative4E1560 preserves GAME.EXE 004E1560 while keeping
// all transient object, modifier, update-data, and damage-callback pointers at
// the host pointer width.
func PlayerDamageWeaponNative4E1560(
	item, owner, source, effective *Object,
	amount float32,
	typ object.DamageType,
	runtime ItemDurabilityDamageRuntime4E1560,
) bool {
	return itemDurabilityDamage4E1560(item, owner, source, effective, amount, typ, true, runtime)
}

// EquipDamageNative4E16D0 is the equipped-armor counterpart of 004E1560. It
// intentionally does not apply the weapon subclass exclusion.
func EquipDamageNative4E16D0(
	item, owner, source, effective *Object,
	amount float32,
	typ object.DamageType,
	runtime ItemDurabilityDamageRuntime4E1560,
) bool {
	return itemDurabilityDamage4E1560(item, owner, source, effective, amount, typ, false, runtime)
}
