package server

import "github.com/opennox/libs/object"

// Script Damage may retain a nonunit world object as both source and weapon.
// No unit update record exists. Already ported specialized collider slices
// retain their existing admission and partial-runtime contracts.
func playerDamageWorldSelfShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	if source == nil || source != weapon || source.Class().HasAny(object.MaskUnits|object.ClassWeapon|object.ClassWand|object.ClassMissile) {
		return false
	}
	if playerDamageWorldFlameShape4E17B0(weapon, typ) || playerDamageCloudPoisonShape4E17B0(weapon, typ) || playerDamageWorldImpaleShape4E17B0(source, weapon, typ) || playerDamageNPCTurnUndeadShape4E17B0(source, weapon, typ) {
		return false
	}
	if typ == object.DamageZapRay && source.Class().Has(object.ClassSimple) && source.Class().Has(object.ClassImmobile) {
		return false
	}
	return true
}
