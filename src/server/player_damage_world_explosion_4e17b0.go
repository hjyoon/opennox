package server

import "github.com/opennox/libs/object"

// Both stock BlackPowderBarrel*Breaking types are SIMPLE|LIGHT and have
// no unit update data. 0053C9A0 passes that object as source, with a nil
// weapon, to case 7. 004E0C55 skips the melee gate when weapon is nil;
// 004E1ABA uses source.PrevPos without dereferencing source.UpdateData.
// Keep this world-source admission separate from unit/missile explosions.
func playerDamageWorldExplosionShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	return source != nil && weapon == nil && typ == object.DamageExplosion &&
		source.Class().Has(object.ClassSimple) &&
		!source.Class().HasAny(object.MaskUnits|object.ClassMissile|object.ClassWeapon|object.ClassWand)
}
