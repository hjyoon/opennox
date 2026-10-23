package server

import "github.com/opennox/libs/object"

// Initialized wands retain the complete prefix and modifier tail. Native
// WandUseData owns the live melee flag whenever the subclass selects it.
func playerDamageInitializedWandShape4E17B0(source, weapon *Object) bool {
	if weapon == nil || weapon.InitData == nil || !weapon.Class().Has(object.ClassWand) || weapon.Class().HasAny(object.MaskUnits|object.ClassMissile) {
		return false
	}
	if uint32(weapon.SubClass())&0x047f0000 != 0 && weapon.UseData.Ptr == nil {
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
