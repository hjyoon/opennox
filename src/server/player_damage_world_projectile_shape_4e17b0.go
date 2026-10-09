package server

import "github.com/opennox/libs/object"

// SkullUpdate supplies the terminal world owner, not a unit update record.
// ArrowTrap uses MISSILE|WEAPON arrows; Skull uses a pure spell missile.
// DeathBall/Fragment use CRUSH for both direct and weapon-less nearby hits.
// Keep these admissions disjoint from melee, wand and unit-shaped weapons.
func playerDamageWorldProjectileShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	if source == nil {
		return false
	}
	world := source.Class().HasAny(object.ClassImmobile|object.ClassSimple) &&
		!source.Class().HasAny(object.MaskUnits|object.ClassWeapon|object.ClassWand|object.ClassMissile)
	pureMissile := func(o *Object) bool {
		return o != nil && o.Class().Has(object.ClassMissile) &&
			!o.Class().HasAny(object.MaskUnits|object.ClassWeapon|object.ClassWand)
	}
	switch typ {
	case object.DamageImpale:
		return world && source.Class().Has(object.ClassImmobile) && source != weapon && weapon != nil && weapon.Class().Has(object.ClassMissile) &&
			!weapon.Class().HasAny(object.MaskUnits|object.ClassWand) &&
			!defaultDamageAttackQualifies4E1400(source, weapon)
	case object.DamageFlame:
		return world && source.Class().Has(object.ClassImmobile) && source != weapon && pureMissile(weapon)
	case object.DamageCrush:
		if weapon == nil {
			return pureMissile(source)
		}
		return pureMissile(weapon) && (source == weapon || source.Class() == 0 || world ||
			(source.UpdateData != nil && source.Class().HasAny(object.MaskUnits) &&
				!source.Class().HasAny(object.ClassWeapon|object.ClassWand|object.ClassMissile)))
	}
	return false
}
