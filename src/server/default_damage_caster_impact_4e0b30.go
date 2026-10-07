package server

import "github.com/opennox/libs/object"

// defaultDamageCasterImpactShape4E0B30 admits the non-weapon IMPACT tail
// used by 0052DEC0. Earthquake passes the terminal owner as source and the
// complete caster as weapon: those can be the same PLAYER, an owned MONSTER,
// or a class-zero ImaginaryCaster. None is a projectile or a melee item.
// Hostility and Shock still use the live 004E1400/004E1470 predicates; do not
// infer their result from this admission predicate or require UpdateData for
// a class that GAME.EXE never dereferences on the executed path.
func defaultDamageCasterImpactShape4E0B30(weapon *Object, typ object.DamageType) bool {
	return typ == object.DamageImpact && weapon != nil &&
		!weapon.Class().HasAny(object.ClassWeapon|object.ClassWand|object.ClassMissile)
}
