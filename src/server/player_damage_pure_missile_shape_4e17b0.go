package server

import "github.com/opennox/libs/object"

// Use the complete original prefix for initialized pure-missile forms.
// Partial runtime callers and the established SIMPLE/Pixie path retain
// their earlier entry and live-callback contracts.
func playerDamagePureMissileShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	if typ == object.DamageImpact && weapon != nil && weapon.Class().Has(object.ClassSimple) && playerDamageSpellMissileImpactShape4E17B0(source, weapon, typ) {
		return false
	}
	return playerDamagePureMissileSourceShape4E17B0(source, weapon)
}

// DefaultDamage accepts the same initialized source and missile independent
// of which PlayerDamage prefix handled it. No source/missile update exists
// for nonunits; an unowned missile can be its own terminal source.
func playerDamagePureMissileSourceShape4E17B0(source, weapon *Object) bool {
	if weapon == nil || !weapon.Class().Has(object.ClassMissile) || weapon.Class().HasAny(object.MaskUnits|object.ClassWeapon|object.ClassWand) {
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
