package server

import "github.com/opennox/libs/object"

// SparkExplosionCollide supplies the owning unit and a distinct missile for
// direct FLAME hits, but MapDamageUnits supplies the missile and nil weapon for
// splash. Neither shape qualifies for the original melee/Shock predicate.
func playerDamageMissileFlameShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	if typ != object.DamageFlame || source == nil {
		return false
	}
	missile := source
	if weapon != nil {
		if source == weapon || source.UpdateData == nil || !source.Class().HasAny(object.ClassPlayer|object.ClassMonster) {
			return false
		}
		missile = weapon
	}
	return missile.Class().Has(object.ClassMissile) &&
		!missile.Class().HasAny(object.MaskUnits|object.ClassWeapon|object.ClassWand)
}
