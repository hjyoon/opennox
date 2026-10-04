package server

import "github.com/opennox/libs/object"

// Troll's ordinary strike passes the monster as BOTH source and weapon.
// Case 11 shares the full-armor/carry switch at 004E1F84; the separate
// 004E0C55 friendly-hit/004E0C9E Shock gates still qualify this self-weapon.
// This is not a ranged projectile, nor a player Berserker Charge.
func playerDamageMonsterImpactShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	return typ == object.DamageImpact && source != nil && source == weapon &&
		source.UpdateData != nil && source.Class().Has(object.ClassMonster) &&
		!source.Class().HasAny(object.ClassPlayer|object.ClassWeapon|object.ClassWand|object.ClassMissile)
}
