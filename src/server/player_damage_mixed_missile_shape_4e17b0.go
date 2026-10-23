package server

import "github.com/opennox/libs/object"

// Non-melee weapon-class missiles retain the same prefix as other projectiles, then
// enter original item pre-damage through their initialized modifier record.
// The missile need not carry unit update data; source unit records do.
func playerDamageMixedMissileShape4E17B0(source, weapon *Object) bool {
	if source == nil || weapon == nil || weapon.InitData == nil || (!weapon.Class().Has(object.ClassMissile) || !weapon.Class().Has(object.ClassWeapon)) || weapon.Class().HasAny(object.MaskUnits|object.ClassWand) {
		return false
	}
	if defaultDamageAttackQualifies4E1400(source, weapon) {
		return false
	}
	if source == weapon {
		return true
	}
	if source.Class().HasAny(object.ClassWeapon | object.ClassWand | object.ClassMissile) {
		return false
	}
	if source.Class().Has(object.ClassPlayer) {
		return !source.Class().Has(object.ClassMonster) && source.UpdateData != nil && source.UpdateDataPlayer().Player != nil
	}
	return !source.Class().Has(object.ClassMonster) || source.UpdateData != nil
}
