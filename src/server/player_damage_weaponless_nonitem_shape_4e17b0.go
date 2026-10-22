package server

import "github.com/opennox/libs/object"

// A direct damage callback can retain a non-item source with no weapon.
// The prefix uses the source for facing/Reflect, but every damage and wear
// callback must retain the original nil weapon argument.
func playerDamageWeaponlessNonItemShape4E17B0(source, weapon *Object, _ object.DamageType) bool {
	if source == nil || weapon != nil || source.Class().HasAny(object.ClassWeapon|object.ClassWand|object.ClassMissile) {
		return false
	}
	if source.Class().Has(object.ClassPlayer) {
		return !source.Class().Has(object.ClassMonster) && source.UpdateData != nil && source.UpdateDataPlayer().Player != nil
	}
	return !source.Class().Has(object.ClassMonster) || source.UpdateData != nil
}
