package server

import "github.com/opennox/libs/object"

// DamageCollide supplies a world hazard and its terminal parent. Stock Spike
// is IMMOBILE, not SIMPLE; SpikeBlock is SIMPLE, and its immobile variant is
// not. Their common admission is DANGEROUS/non-unit IMPALE (case 3), without a
// type ID or positive-only gate. Unit/weapon/wand/missile shapes stay disjoint.
func playerDamageWorldImpaleShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	if typ != object.DamageImpale || source == nil || weapon == nil ||
		!weapon.Class().Has(object.ClassDangerous) ||
		weapon.Class().HasAny(object.MaskUnits|object.ClassWeapon|object.ClassWand|object.ClassMissile) {
		return false
	}
	return source == weapon || (source.UpdateData != nil && source.Class().HasAny(object.MaskUnits) &&
		!source.Class().HasAny(object.ClassWeapon|object.ClassWand|object.ClassMissile))
}
