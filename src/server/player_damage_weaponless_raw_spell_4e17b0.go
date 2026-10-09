package server

import "github.com/opennox/libs/object"

// Death Ray and Mana Bomb's radial calls supply a terminal unit owner, nil,
// or the class-zero ImaginaryCaster, and no weapon. They are not the electric
// armor/protection cases 9/17, nor SentryGlobe's nonnil-weapon ray prefix.
func playerDamageWeaponlessRawSpellShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	return weapon == nil && (typ == object.DamageManaBomb || typ == object.DamageZapRay) &&
		(source == nil || source.Class() == 0 || (source.Class().HasAny(object.MaskUnits) &&
			!source.Class().HasAny(object.ClassWeapon|object.ClassWand|object.ClassMissile)))
}
