package server

import "github.com/opennox/libs/object"

// Terminal non-item owners need not have the same class or address as the
// actual attack. Unit records must be initialized; world records need none.
func playerDamageDistinctNonItemShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	if source == nil || weapon == nil || source == weapon {
		return false
	}
	valid := func(obj *Object) bool {
		class := obj.Class()
		if class.HasAny(object.ClassWeapon | object.ClassWand | object.ClassMissile) {
			return false
		}
		if class.Has(object.ClassPlayer) {
			return !class.Has(object.ClassMonster) && obj.UpdateData != nil && obj.UpdateDataPlayer().Player != nil
		}
		return !class.Has(object.ClassMonster) || obj.UpdateData != nil
	}
	if !valid(source) || !valid(weapon) || playerDamageOwnedMonsterShape4E17B0(source, weapon, typ) {
		return false
	}
	// Keep previously restored collider and unit-owned ray contracts.
	if playerDamageWorldFlameShape4E17B0(weapon, typ) || playerDamageCloudPoisonShape4E17B0(weapon, typ) || playerDamageWorldImpaleShape4E17B0(source, weapon, typ) || playerDamageNPCTurnUndeadShape4E17B0(source, weapon, typ) {
		return false
	}
	if typ == object.DamageZapRay && weapon.Class().Has(object.ClassSimple) && weapon.Class().Has(object.ClassImmobile) && source.Class().HasAny(object.MaskUnits) {
		return false
	}
	return true
}
