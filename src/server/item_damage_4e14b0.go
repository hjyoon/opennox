package server

import "github.com/opennox/libs/object"

// ItemDamageDefault4E14B0 is the native-width form of the default-damage
// callback used by GAME.EXE 004E14B0 and 004E1500.
type ItemDamageDefault4E14B0 func(
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
) bool

// BallDamage4E14A0 preserves the inert GAME.EXE 004E14A0 callback.
func BallDamage4E14A0(
	_, _, _ *Object,
	_ int32,
	_ object.DamageType,
) bool {
	return false
}

// WeaponDamage4E14B0 preserves GAME.EXE 004E14B0 without narrowing Object
// pointers to the original PE32 int arguments. Weapons and wands are damaged
// while held; loose items accept only LAVA damage.
func WeaponDamage4E14B0(
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
	applyDefault ItemDamageDefault4E14B0,
) bool {
	if target == nil ||
		!target.ObjClass.HasAny(object.ClassWeapon|object.ClassWand) ||
		(target.InvHolder == nil && typ != object.DamageLava) ||
		applyDefault == nil {
		return false
	}
	return applyDefault(target, source, weapon, damage, typ)
}

// ArmorDamage4E1500 preserves GAME.EXE 004E1500 without narrowing Object
// pointers to the original PE32 int arguments. Loose armor accepts only LAVA
// damage, and CRUSH damage against metal armor is doubled with int32 wrapping.
func ArmorDamage4E1500(
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
	applyDefault ItemDamageDefault4E14B0,
) bool {
	// The PE32 body dereferences target unconditionally. A nil target cannot be
	// a legitimate damage-dispatch receiver, so the native-width bridge rejects
	// it instead of reproducing an invalid-memory access.
	if target == nil ||
		!target.ObjClass.Has(object.ClassArmor) ||
		(target.InvHolder == nil && typ != object.DamageLava) {
		return false
	}
	if typ == object.DamageCrush && object.Material(target.Material).Has(object.MaterialMetal) {
		damage *= 2
	}
	if damage == 0 || applyDefault == nil {
		return false
	}
	return applyDefault(target, source, weapon, damage, typ)
}
