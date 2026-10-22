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
		// A terminal ranged arrow can itself be both source and weapon.
		// Keep melee/unit/wand shapes outside the no-Shock projectile tail.
		if source == weapon && source.Class().Has(object.ClassMissile) && source.Class().Has(object.ClassWeapon) &&
			!source.Class().HasAny(object.MaskUnits|object.ClassWand) && !defaultDamageAttackQualifies4E1400(source, weapon) {
			return true
		}
		return world && source.Class().Has(object.ClassImmobile) && source != weapon && weapon != nil && weapon.Class().Has(object.ClassMissile) &&
			!weapon.Class().HasAny(object.MaskUnits|object.ClassWand) &&
			!defaultDamageAttackQualifies4E1400(source, weapon)
	case object.DamageFlame, object.DamageExplosion:
		// A terminal Fireball or Magic Missile owner can be the missile itself. Preserve both
		// arguments through the world-projectile prefix: no unit update exists,
		// and 004E1AA2 leaves the marker clear until case 1/7 records the type.
		if source == weapon && pureMissile(weapon) {
			return true
		}
		// Scripted spell missiles can retain a class-zero ImaginaryCaster.
		return (world && source.Class().Has(object.ClassImmobile) || source.Class() == 0) &&
			source != weapon && pureMissile(weapon)
	case object.DamageCrush:
		// PressurePlate creates a distinct SIMPLE LargeFist, not a missile.
		if world && source != weapon && weapon != nil && weapon.Class().Has(object.ClassSimple) &&
			!weapon.Class().HasAny(object.MaskUnits|object.ClassWeapon|object.ClassWand|object.ClassMissile) {
			return true
		}
		if weapon == nil {
			return pureMissile(source)
		}
		return pureMissile(weapon) && (source == weapon || source.Class() == 0 || world ||
			(source.UpdateData != nil && source.Class().HasAny(object.MaskUnits) &&
				!source.Class().HasAny(object.ClassWeapon|object.ClassWand|object.ClassMissile)))
	}
	return false
}

// UndeadKiller is LIGHT|CLIENT_PREDICT, not a missile or melee weapon.
// The collision supplies a terminal owner or the collider itself; direct
// callbacks may also retain a nil source with this distinct collider.
func playerDamageNPCTurnUndeadShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	if typ != object.DamageDispelUndead || weapon == nil || !weapon.Class().Has(object.ClassLight) || !weapon.Class().Has(object.ClassClientPredict) || weapon.Class().HasAny(object.MaskUnits|object.ClassWeapon|object.ClassWand|object.ClassMissile) {
		return false
	}
	return source == nil || (!source.Class().HasAny(object.ClassWeapon|object.ClassWand|object.ClassMissile) && (!source.Class().HasAny(object.MaskUnits) || source.UpdateData != nil))
}

// Complete the initialized-player source/weapon forms of the fire switch.
func playerDamagePlayerSelfFireShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	if typ != object.DamageFlame && typ != object.DamageExplosion && typ != object.DamageLava {
		return false
	}
	return source != nil && source == weapon && source.UpdateData != nil && source.Class().Has(object.ClassPlayer) && !source.Class().HasAny(object.ClassMonster|object.ClassWeapon|object.ClassWand|object.ClassMissile) && source.UpdateDataPlayer().Player != nil
}
func playerDamagePlayerSelfImpactShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	return typ == object.DamageImpact && source != nil && source == weapon && source.UpdateData != nil && source.Class().Has(object.ClassPlayer) && !source.Class().HasAny(object.ClassMonster|object.ClassWeapon|object.ClassWand|object.ClassMissile) && source.UpdateDataPlayer().Player != nil
}

// The same three fire cases accept terminal script monster sources.
func playerDamageMonsterSelfFireShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	if typ != object.DamageFlame && typ != object.DamageExplosion && typ != object.DamageLava {
		return false
	}
	return source != nil && source == weapon && source.UpdateData != nil && source.Class().Has(object.ClassMonster) && !source.Class().HasAny(object.ClassPlayer|object.ClassWeapon|object.ClassWand|object.ClassMissile)
}
