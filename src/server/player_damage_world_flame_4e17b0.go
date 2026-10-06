package server

import "github.com/opennox/libs/object"

// Stock Flame and red FlameCleanse are world fires, not spell missiles. Their
// colliders can supply a nil, self, imaginary or terminal unit parent. Only
// case 1 is admitted here; weapon/wand/unit/missile damage keeps its own path.
func playerDamageWorldFlameShape4E17B0(weapon *Object, typ object.DamageType) bool {
	return typ == object.DamageFlame && weapon != nil &&
		weapon.Class().Has(object.ClassFire|object.ClassSimple|object.ClassDangerous) &&
		!weapon.Class().HasAny(object.MaskUnits|object.ClassWeapon|object.ClassWand|object.ClassMissile)
}
