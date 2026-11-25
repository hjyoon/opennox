package server

import "github.com/opennox/libs/object"

// A direct damage callback can retain a non-item source with no weapon.
// The prefix uses the source for facing/Reflect, but every damage and wear
// callback must retain the original nil weapon argument.
func playerDamageWeaponlessNonItemShape4E17B0(source, weapon *Object, _ object.DamageType) bool {
	if source == nil || weapon != nil {
		return false
	}
	class := source.Class()
	if class.HasAny(object.ClassWeapon|object.ClassWand) ||
		(class.Has(object.ClassMissile) && class.HasAny(object.MaskUnits)) {
		return false
	}
	if class.Has(object.ClassPlayer) {
		return !class.Has(object.ClassMonster) && source.UpdateData != nil && source.UpdateDataPlayer().Player != nil
	}
	return !class.Has(object.ClassMonster) || source.UpdateData != nil
}
