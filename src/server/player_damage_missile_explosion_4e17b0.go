package server

import "github.com/opennox/libs/object"

// BoomCollide sends a distinct Magic Missile and its owning unit for the
// direct hit; MapDamageUnits sends the missile with a nil weapon for splash.
// Pure spell missiles do not enter the original melee/Shock predicate.
func playerDamageMissileExplosionShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	if typ != object.DamageExplosion || source == nil {
		return false
	}
	missile := source
	if weapon != nil {
		if source == weapon || source.UpdateData == nil || !source.Class().HasAny(object.ClassPlayer|object.ClassMonster) {
			return false
		}
		missile = weapon
	}
	return missile.Class().Has(object.ClassMissile) && !missile.Class().HasAny(object.MaskUnits|object.ClassWeapon|object.ClassWand)
}
