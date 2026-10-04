package server

import "github.com/opennox/libs/object"

// PixieCollide (004EA080) supplies its live terminal parent and the Pixie
// itself as weapon, with IMPACT/type 11. An unowned Pixie is its own parent.
// This admission describes pure spell missiles, not a cached Pixie type ID;
// weapon/wand/unit shapes keep their separate melee, Shock and modifier paths.
func playerDamageSpellMissileImpactShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	if typ != object.DamageImpact || source == nil || weapon == nil ||
		!weapon.Class().Has(object.ClassMissile) ||
		weapon.Class().HasAny(object.MaskUnits|object.ClassWeapon|object.ClassWand) {
		return false
	}
	return source == weapon || (source.UpdateData != nil &&
		source.Class().HasAny(object.MaskUnits) &&
		!source.Class().HasAny(object.ClassWeapon|object.ClassWand|object.ClassMissile))
}
