package server

import "github.com/opennox/libs/object"

// Initialized weapons without missile/wand/unit bits share the complete native
// prefix and item pre-damage tail. Their live subclass still controls friendly
// attacks and Shock; the actual source identity, including nil, is retained.
func playerDamageInitializedWeaponShape4E17B0(source, weapon *Object) bool {
	if weapon == nil || weapon.InitData == nil || !weapon.Class().Has(object.ClassWeapon) || weapon.Class().HasAny(object.MaskUnits|object.ClassWand|object.ClassMissile) {
		return false
	}
	if source == nil || source == weapon {
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
