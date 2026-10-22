package server

import "github.com/opennox/libs/object"

// Complete the remaining physical, raw/default and electric types for a
// pure spell missile. Existing flame/crush/pierce/explosion and unit/self-owned SIMPLE impact prefixes
// retain their specialized admission and partial-runtime contracts.
func playerDamagePureMissileShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	switch typ {
	case object.DamageFlame, object.DamageCrush, object.DamageImpale, object.DamageExplosion:
		return false
	}
	if typ == object.DamageImpact && weapon != nil && weapon.Class().Has(object.ClassSimple) && playerDamageSpellMissileImpactShape4E17B0(source, weapon, typ) {
		return false // Existing unit/self-owned Pixie prefix and its live-callback contracts.
	}
	if source == nil || weapon == nil || !weapon.Class().Has(object.ClassMissile) || weapon.Class().HasAny(object.MaskUnits|object.ClassWeapon|object.ClassWand) {
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
